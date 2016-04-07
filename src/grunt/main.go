package main

import (
	"flag"
	"fmt"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type SMTP struct {
	From     string
	Username string
	Password string
	Server   string
	Port     int
}

type Config struct {
	Services        []*Service          `json:"services"`
	ServiceMap      map[string]*Service `json:omit`
	Mail            SMTP
	Server          string
	Directory       string
	ConfigDirectory string `yaml:"configDirectory"`
}

var config Config

func main() {
	var port int
	flag.IntVar(&port, "p", 9901, "specify port to use.  defaults to 9901.")

	config.ServiceMap = make(map[string]*Service)
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("Usage: grunt gruntfile.yml")
	}
	gruntfile := flag.Arg(0)
	data, err := ioutil.ReadFile(gruntfile)
	if err != nil {
		log.Fatal("Error reading %v: %v", gruntfile, err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error in YML parsing: %v", err)
	}

	// Read all the files in the config directory
	if config.ConfigDirectory != "" {
		log.Printf("load configurations from %v", config.ConfigDirectory)
		loadServices(config.ConfigDirectory)
	}

	// Start up all the services
	for _, service := range config.Services {
		service.setup()
		config.ServiceMap[service.EndPoint] = service
		log.Printf("\tservice available: %v\n", service.EndPoint)
	}
	if config.Mail.Port == 0 {
		config.Mail.Port = 25
	}
	if config.Mail.From == "" {
		config.Mail.From = "noreply@example.com"
	}
	if config.Directory == "" {
		config.Directory, err = ioutil.TempDir("", "grunt")
		if err != nil {
			log.Fatalf("Failed to make working directory(%v): %v", config.Directory, err.Error())
		}
	}
	err = os.MkdirAll(config.Directory, 0755)
	if err != nil {
		log.Fatalf("Failed to make working directory(%v): %v", config.Directory, err.Error())
	}

	// Expose the endpoints
	r := mux.NewRouter()
	r.HandleFunc("/rest/service", GetServices).Methods("GET")
	r.HandleFunc("/rest/service/{id}", GetService).Methods("GET")
	r.HandleFunc("/rest/service/{id}", StartService).Methods("POST")
	r.HandleFunc("/rest/job/{id}", GetJob).Methods("GET")
	r.HandleFunc("/rest/job/wait/{id}", WaitForJob).Methods("GET")
	r.HandleFunc("/rest/job/{id}/file/{filename}", GetJobFile).Methods("GET")

	r.HandleFunc("/help.html", Help).Methods("GET")
	r.HandleFunc("/jobs.html", Jobs).Methods("GET")
	r.HandleFunc("/job/{id}", JobDetail).Methods("GET")
	r.HandleFunc("/services.html", Services).Methods("GET")
	r.HandleFunc("/submit/{id}.html", Submit).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo}))

	http.Handle("/", r)
	log.Printf("Starting grunt on http://localhost:%v", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

}
