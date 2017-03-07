package main

//go:generate bin/go-bindata -prefix assets -o assets.go assets/...

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
	SlicerServices  []*SlicerService    `yaml:"cli" json:"-"`
	ServiceMap      map[string]*Service `json:"-"`
	Mail            SMTP                `json:"mail"`
	Server          string              `json:"server"`
	Directory       string              `json:"working_directory"`
	ConfigDirectory string              `json:"config_directory" yaml:"configDirectory"`
	WarnLevel       int                 `json:"warn_level" yaml:"warnLevel"`
	CriticalLevel   int                 `json:"critical_level" yaml:"criticalLevel"`
}

var config Config
var consulHost string
var consulPort int
var advertisedHost string
var advertisedPort int

func main() {
	var port int
	flag.IntVar(&port, "p", 9901, "specify port to use.  defaults to 9901.")
	flag.StringVar(&consulHost, "consul", "", "specify Consul host. defaults to none. Also set by CONSUL_HOST or CONSULT_PORT_8500_TCP_ADDR environment variable")
	flag.IntVar(&consulPort, "consul-port", 0, "specify Consul port to use.  defaults to 0.  Also set through the CONSULT_HOST or CONSUL_PORT_8500_TCP_PORT environment variable set by Docker")
	flag.StringVar(&advertisedHost, "advertised", "", "specify Advertised host. defaults to none.  Also set through the ADVERTISED_HOST environment variable.")
	flag.IntVar(&advertisedPort, "advertised-port", 0, "specify Advertised port to use.  defaults to 0. Also set through the ADVERTISED_PORT environment variable.")

	// Set config defaults
	config.ServiceMap = make(map[string]*Service)
	config.WarnLevel = 3
	config.CriticalLevel = 5
	config.Mail.Port = 25

	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("Usage: grunt gruntfile.yml")
	}
	gruntfile := flag.Arg(0)
	data, err := ioutil.ReadFile(gruntfile)
	if err != nil {
		log.Fatalf("Error reading %v: %v", gruntfile, err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error in YML parsing: %v", err)
	}

	for _, ss := range config.SlicerServices {
		log.Printf("Found %+v Slicer service\n", *ss)
		s, err := CreateService(ss.Executable)
		if err != nil {
			log.Fatalf("Error constructing Slicer CLI: %v", err)
		}
		s.EndPoint = ss.EndPoint
		config.Services = append(config.Services, s)
	}

	// Register the main grunt services
	c := ConfigD{Name: "grunt", Services: config.Services}
	registerConfigWithConsul(&c)

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
	r.HandleFunc("/rest/job/{id}/zip", GetJobZip).Methods("GET")

	r.HandleFunc("/help.html", Help).Methods("GET")
	r.HandleFunc("/jobs.html", Jobs).Methods("GET")
	r.HandleFunc("/job/{id}", JobDetail).Methods("GET")
	r.HandleFunc("/services.html", Services).Methods("GET")
	r.HandleFunc("/submit/{id}.html", Submit).Methods("GET")
	r.HandleFunc("/health", GetHealth).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo}))

	http.Handle("/", r)
	log.Printf("Starting grunt on http://localhost:%v", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

}
