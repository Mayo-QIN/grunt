package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/imdario/mergo"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

var jobs = make(map[string]*Job)

type Service struct {
	EndPoint    string            `yaml:"endPoint" json:"endPoint"`
	CommandLine []string          `yaml:"commandLine" json:"commandLine"`
	Description string            `json:"description"`
	Defaults    map[string]string `yaml:defaults json:defaults`
}

type Job struct {
	UUID              string            `json:"uuid"`
	CommandLine       []string          `yaml:"commandLine" json:"commandLine"`
	ParsedCommandLine []string          `json:"-"`
	FileMap           map[string]string `json:"-"`
	StartTime         time.Time         `json:"startTime"`
	EndTime           time.Time         `json:"endTime"`
	Status            string
	Address           []string
	Endpoint          string
	// Running process
	cmd    *exec.Cmd
	Output bytes.Buffer
}

func Template(name string, data map[string]interface{}, w http.ResponseWriter, request *http.Request) {
	var templateData = map[string]interface{}{
		"jobs":       jobs,
		"services":   config.Services,
		"serviceMap": config.ServiceMap,
	}
	// merge in our extra data
	mergo.Map(&templateData, data)
	contents, _ := Asset("template/" + name + ".html")
	t, _ := template.New(name).Parse(string(contents))
	t.Execute(w, templateData)
}
func Help(w http.ResponseWriter, request *http.Request) {
	Template("help", nil, w, request)
}
func Jobs(w http.ResponseWriter, request *http.Request) {
	Template("jobs", nil, w, request)
}
func Submit(w http.ResponseWriter, request *http.Request) {
	Template("submit", nil, w, request)
}
func Services(w http.ResponseWriter, request *http.Request) {
	Template("services", nil, w, request)
}
func JobDetail(w http.ResponseWriter, request *http.Request) {
	key := mux.Vars(request)["id"]
	job := jobs[key]
	if job == nil {
		http.Error(w, "could not find job", http.StatusNotFound)
		return
	}
	var data = map[string]interface{}{
		"job": job}
	Template("job", data, w, request)
}

func GetServices(w http.ResponseWriter, request *http.Request) {
	json.NewEncoder(w).Encode(config)
}

func GetService(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	service := config.ServiceMap[vars["id"]]
	json.NewEncoder(w).Encode(service)
}

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
		CommandLine: service.CommandLine,
		FileMap:     make(map[string]string),
		Endpoint:    service.EndPoint,
	}

	// do we have an email address?
	if request.MultipartForm.Value["mail"] != nil {
		job.Address = request.MultipartForm.Value["mail"]
	}

	cl := make([]string, 0)
	// Make a temp directory
	dir, err := ioutil.TempDir("", job.UUID)
	for _, arg := range service.CommandLine {
		log.Printf("Parsing %v", arg)
		// Do we start with an @?
		key := arg[1:]
		prefix := arg[0]
		if prefix == '@' {
			// Lookup first in form
			if request.MultipartForm.Value[key] != nil {
				cl = append(cl, request.MultipartForm.Value[key][0])
			} else {
				// Look up in defaults
				cl = append(cl, service.Defaults[key])
			}
		} else if prefix == '<' {
			// Do we have an < to indicate an uploaded file?
			v := request.MultipartForm.File[key]
			if v == nil {
				http.Error(w, fmt.Sprintf("Could not find %v in form data", key), http.StatusInternalServerError)
				return
			}
			header := v[0]
			// Save a temp file
			fout, err := os.Create(filepath.Join(dir, filepath.Base(header.Filename)))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			f, err := header.Open()
			count, err := io.Copy(fout, f)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("Wrote %v bytes to %v", count, header.Filename)
			fout.Close()
			cl = append(cl, fout.Name())
		} else if prefix == '>' {
			// Write a file...
			// Save a temp file
			if request.MultipartForm.Value[key] == nil {
				http.Error(w, fmt.Sprintf("filename must be specified for %v", key), http.StatusInternalServerError)
				return
			}
			tmp := filepath.Join(dir, filepath.Base(request.MultipartForm.Value[key][0]))
			job.FileMap[key] = tmp
			cl = append(cl, tmp)
		} else {
			cl = append(cl, arg)
		}
	}
	log.Printf("Final command line: %v", cl)
	job.ParsedCommandLine = cl
	cmd := exec.Command(cl[0], cl[1:]...)
	job.StartTime = time.Now()
	cmd.Stdout = &job.Output
	cmd.Stderr = &job.Output
	job.cmd = cmd
	job.Status = "pending"

	// Launch a go routine to wait
	go func() {
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
		// Send email here
		log.Printf("Would send email to %v", job.Address)
		Email(&job)
		// Cleanup after 10 minutes
		<-time.After(time.Minute * 120)
		Cleanup(&job)
	}()

	json.NewEncoder(w).Encode(job)
	jobs[job.UUID] = &job
}

func GetJob(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	job := jobs[vars["id"]]
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
	file := job.FileMap[vars["filename"]]
	w.Header().Set("Content-Disposition", "attachment;filename="+filepath.Base(file))
	http.ServeFile(w, request, file)
}
