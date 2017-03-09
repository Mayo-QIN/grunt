package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

// Return the information about a job as JSON
func JobDetail(w http.ResponseWriter, request *http.Request) {
	key := mux.Vars(request)["id"]
	job := jobs[key]
	if job == nil {
		http.Error(w, "could not find job", http.StatusNotFound)
		return
	}
	var data = map[string]interface{}{"job": job}
	Template("job", data, w, request)
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
		Endpoint:    service.EndPoint,
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
	cmd := exec.Command(job.ParsedCommandLine[0], job.ParsedCommandLine[1:]...)
	cmd.Dir = job.WorkingDirectory
	job.StartTime = time.Now()
	cmd.Stdout = &job.Output
	cmd.Stderr = &job.Output
	job.cmd = cmd
	job.Status = "pending"

	// Launch a go routine to wait
	go func() {
		jobMutex.Lock()
		job.Status = "running"
		job.cmd.Start()
		err := job.cmd.Wait()
		job.EndTime = time.Now()
		if err != nil {
			job.Status = "error"
		} else {
			if job.cmd.ProcessState.Success() {
				job.Status = "success"
			} else {
				job.Status = "failed"
			}
		}
		jobMutex.Unlock()
		// Notify waiters
		log.Printf("%v completed with status %v", job.UUID, job.Status)
		for _, c := range job.waiters {
			c <- true
		}

		// Send email here
		Email(&job)

		// Cleanup after 120 minutes
		<-time.After(time.Minute * 120)
		Cleanup(&job)
	}()

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
	if job.Status == "running" {
		c := make(chan bool)
		job.Lock()
		job.waiters = append(job.waiters, c)
		job.Unlock()
		<-c
		close(c)
	}
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
		w.Header().Set("Content-Disposition", "attachment;filename="+filepath.Base(file))
		http.ServeFile(w, request, file)
		return
	}
	file, ok = job.ZipMap[vars["filename"]]
	if ok {
		outputDirectory := file
		log.Printf("file: %v", file)
		log.Printf("Requested %v: sending directory %v as a zip file", vars["filename"], outputDirectory)
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(file)+"\"")
		gz := zip.NewWriter(w)
		defer gz.Close()

		filepath.Walk(outputDirectory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// Take off the path, and a leading "/" to make the output relative
			header.Name = strings.TrimPrefix(path, file)
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

	w.Header().Set("Content-Disposition", "attachment;filename="+job.Endpoint+"-"+job.UUID+".zip")
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
