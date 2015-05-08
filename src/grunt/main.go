package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	graceful "gopkg.in/tylerb/graceful.v1"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Services   []*Service          `json:"services"`
	ServiceMap map[string]*Service `json:omit`
}

var config Config

func main() {
	config.ServiceMap = make(map[string]*Service)

	if len(os.Args) < 2 {
		log.Fatal("Usage: grunt gruntfile.yml")
	}
	gruntfile := os.Args[1]
	data, err := ioutil.ReadFile(gruntfile)
	if err != nil {
		log.Fatal("Error reading %v: %v", gruntfile, err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error in YML parsing: %v", err)
	}
	for _, service := range config.Services {
		config.ServiceMap[service.EndPoint] = service
	}

	// Expose the endpoints
	r := mux.NewRouter()
	r.HandleFunc("/rest/service", GetServices).Methods("GET")
	r.HandleFunc("/rest/service/{id}", GetService).Methods("GET")
	r.HandleFunc("/rest/service/{id}", StartService).Methods("POST")
	r.HandleFunc("/rest/job/{id}", GetJob).Methods("GET")
	r.HandleFunc("/rest/job/{id}/file/{filename}", GetJobFile).Methods("GET")

	http.Handle("/", r)
	// log.Fatal(http.ListenAndServe(":9901", nil))
	log.Info("Starting grunt on port 9901 (http://localhost:9901)")
	graceful.Run(":9901", 10*time.Second, nil)
}
