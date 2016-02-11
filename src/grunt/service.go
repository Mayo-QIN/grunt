package main

import (
	"encoding/json"
	"fmt"
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
	Output            string            `json:"output"`
	StartTime         time.Time         `json:"startTime"`
	EndTime           time.Time         `json:"endTime"`
}

func Template(name string, w http.ResponseWriter, request *http.Request) {
	var templateData = map[string]interface{}{
		"jobs":       jobs,
		"services":   config.Services,
		"serviceMap": config.ServiceMap,
	}
	data, _ := Asset("template/" + name + ".html")
	t, _ := template.New(name).Parse(string(data))
	t.Execute(w, templateData)
}
func Help(w http.ResponseWriter, request *http.Request) {
	Template("help", w, request)
}
func Jobs(w http.ResponseWriter, request *http.Request) {
	Template("jobs", w, request)
}
func Submit(w http.ResponseWriter, request *http.Request) {
	Template("submit", w, request)
}
func Services(w http.ResponseWriter, request *http.Request) {
	Template("services", w, request)
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
	o, err := cmd.Output()
	job.Output = string(o)
	log.Printf("Output... %v", job.Output)
	job.EndTime = time.Now()
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
