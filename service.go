package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

var jobMutex sync.Mutex
var jobs = make(map[string]*Job)

type SlicerService struct {
	EndPoint   string `yaml:"endPoint"`
	Executable string `yaml:"executable"`
}

type Service struct {
	EndPoint          string            `yaml:"endPoint" json:"end_point"`
	CommandLine       []string          `yaml:"commandLine" json:"command_line"`
	Description       string            `json:"description"`
	Defaults          map[string]string `json:"defaults" yaml:"defaults"`
	CreateEmptyOutput bool              `json:"create_empty_output" yaml:"create_empty_output"`
	Arguments         []string          `json:"arguments"`
	Parameters        []string          `json:"parameters"`
	InputFiles        []string          `json:"input_files"`
	OutputFiles       []string          `json:"output_files"`
	InputZip          []string          `json:"input_directories"`
	OutputZip         []string          `json:"output_directories"`
}

// Parse our argements
func NewService() *Service {
	var service Service
	service.Defaults = make(map[string]string)
	service.Arguments = make([]string, 0)
	service.Parameters = make([]string, 0)
	service.InputFiles = make([]string, 0)
	service.OutputFiles = make([]string, 0)
	return &service
}

func (service *Service) setup() *Service {
	if service.Defaults == nil {
		service.Defaults = make(map[string]string)
	}
	if service.Arguments == nil {
		service.Arguments = make([]string, 0)
	}
	if service.Parameters == nil {
		service.Parameters = make([]string, 0)
	}
	if service.InputFiles == nil {
		service.InputFiles = make([]string, 0)
	}
	if service.OutputFiles == nil {
		service.OutputFiles = make([]string, 0)
	}
	if service.InputZip == nil {
		service.InputZip = make([]string, 0)
	}
	if service.OutputZip == nil {
		service.OutputZip = make([]string, 0)
	}
	for _, arg := range service.CommandLine {
		// Do we start with an #?
		key := arg[1:]
		prefix := arg[0]
		isArg := false
		if prefix == '#' {
			isArg = true
			service.Parameters = append(service.Arguments, key)
		} else if prefix == '<' {
			isArg = true
			service.InputFiles = append(service.InputFiles, key)
		} else if prefix == '>' {
			isArg = true
			service.OutputFiles = append(service.OutputFiles, key)
		} else if prefix == '^' {
			isArg = true
			service.InputZip = append(service.InputZip, key)
		} else if prefix == '~' {
			isArg = true
			service.OutputZip = append(service.OutputZip, key)
		}
		if isArg {
			service.Arguments = append(service.Arguments, key)
		}
	}
	return service
}

// Return all services as JSON
func GetServices(w http.ResponseWriter, request *http.Request) {
	json.NewEncoder(w).Encode(config)
}

// Particular info about a service as JSON
func GetService(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	service := config.ServiceMap[vars["id"]]
	json.NewEncoder(w).Encode(service)
}

// Start up the service
func StartService(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	service := config.ServiceMap[vars["id"]]
	if service == nil {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	log.Printf("Found service %v:%v", vars["id"], service)
	// Pull out our arguments
	err := request.ParseMultipartForm(10 * 1024 * 1024)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	job := Job{
		UUID:        uuid.NewV4().String(),
		Service:     service,
		CommandLine: service.CommandLine,
		FileMap:     make(map[string]string),
		ZipMap:      make(map[string]string),
		EndPoint:    service.EndPoint,
		Host:        advertisedHost,
		Port:        advertisedPort,
	}

	// do we have an email address?
	if request.MultipartForm.Value["mail"] != nil {
		job.Address = request.MultipartForm.Value["mail"]
	}

	// Make a working directory
	job.WorkingDirectory = filepath.Join(config.Directory, service.EndPoint, job.UUID)
	err = os.MkdirAll(job.WorkingDirectory, 0755)
	if err != nil {
		log.Printf("Error making working directory: %v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = job.ParseCommandLine(request)
	if err != nil {
		log.Printf("Error parsing command line: %v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	job.Start()
	json.NewEncoder(w).Encode(&job)
	jobs[job.UUID] = &job
}

func GetJob(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	job := jobs[vars["id"]]
	json.NewEncoder(w).Encode(job)
}

func WaitForJob(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	job := jobs[vars["id"]]
	job.Wait()
	json.NewEncoder(w).Encode(job)
}

func GetJobFile(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["id"]
	job := jobs[key]
	if job == nil {
		http.Error(w, fmt.Sprintf("job %v does not exist", key), http.StatusInternalServerError)
		return
	}
	file, ok := job.FileMap[vars["filename"]]
	if ok {
		fileToSend := filepath.Join(job.WorkingDirectory, file)
		log.Printf("Request for %s, sending %s", vars["filename"], fileToSend)
		w.Header().Set("Content-Disposition", "attachment;filename="+filepath.Base(file))
		http.ServeFile(w, request, fileToSend)
		return
	}
	file, ok = job.ZipMap[vars["filename"]]
	if ok {
		outputDirectory := filepath.Join(job.WorkingDirectory, file)
		zipFilename := filepath.Base(vars["filename"])
		if !strings.HasSuffix(zipFilename, ".zip") {
			zipFilename += ".zip"
		}
		log.Printf("Requested %v: sending directory %v as a zip file", vars["filename"], outputDirectory)
		w.Header().Set("Content-Disposition", "attachment; filename=\""+zipFilename+"\"")
		gz := zip.NewWriter(w)
		defer gz.Close()

		filepath.Walk(outputDirectory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Don't create an entry for the root of the directory
			if path == outputDirectory {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// Take off the path, and a leading "/" to make the output relative
			header.Name = strings.TrimPrefix(path, outputDirectory)
			header.Name = strings.TrimPrefix(header.Name, "/")

			// log.Printf("Adding %v from %v", header.Name, path)

			if info.IsDir() {
				header.Name += "/"
			} else {
				header.Method = zip.Deflate
			}

			writer, err := gz.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			return err
		})
		return
	}
}

func GetJobZip(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["id"]
	job := jobs[key]
	if job == nil {
		http.Error(w, fmt.Sprintf("job %v does not exist", key), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", "attachment;filename="+job.EndPoint+"-"+job.UUID+".zip")
	gz := zip.NewWriter(w)
	defer gz.Close()

	filepath.Walk(job.WorkingDirectory, getDirectoryStreamer(job, gz))
}

func getDirectoryStreamer(job *Job, gz *zip.Writer) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		log.Printf("Walking %v", path)
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.Join(job.UUID, strings.TrimPrefix(path, job.WorkingDirectory))

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := gz.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	}
}

func GetHealth(w http.ResponseWriter, request *http.Request) {
	numberOfJobs := len(jobs)
	if numberOfJobs <= config.WarnLevel {
		w.WriteHeader(200)
	} else if numberOfJobs <= config.CriticalLevel {
		w.WriteHeader(429)
	} else {
		w.WriteHeader(500)
	}
	json.NewEncoder(w).Encode(jobs)
}
