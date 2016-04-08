package main

import (
	"fmt"
	consulclient "github.com/hashicorp/consul/api"
	"github.com/satori/go.uuid"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func registerWithConsul() {
	var err error
	// Check environment variables
	if advertisedHost == "" {
		advertisedHost = os.Getenv("ADVERTISED_HOST")
	}
	if advertisedPort == 0 && os.Getenv("ADVERTISED_PORT") != "" {
		advertisedPort, err = strconv.Atoi(os.Getenv("ADVERTISED_PORT"))
		if err != nil {
			log.Printf("Could not parse ADVERTISED_PORT: %v", err.Error())
			return
		}
	}
	if consulHost == "" {
		consulHost = os.Getenv("CONSUL_HOST")
	}
	if consulHost == "" {
		consulHost = os.Getenv("CONSUL_PORT_8500_TCP_ADDR")
	}

	portString := ""
	if consulPort == 0 {
		if os.Getenv("CONSUL_PORT") != "" {
			portString = os.Getenv("CONSUL_PORT")
		}
		if os.Getenv("CONSUL_PORT_8500_TCP_PORT") != "" {
			portString = os.Getenv("CONSUL_PORT_8500_TCP_PORT")
		}
		consulPort, err = strconv.Atoi(portString)
		if err != nil {
			log.Printf("Could not parse port: %v", err.Error())
			return
		}
	}

	log.Printf("Consul: %v:%d", consulHost, consulPort)
	log.Printf("Advertised: %v:%d", advertisedHost, advertisedPort)

	if !(advertisedHost != "" && advertisedPort != 0 && consulHost != "" && consulPort != 0) {
		log.Printf("not connecting to Consul.  Please set consulHost, consulPort, advertisedHost, advertisedPort in gruntfile.yml, or set the CONSUL_ADDR, CONSUL_PORT_8500_TCP_PORT, ADVERTISED_PORT, ADVERTISED_HOST environment variables")
		return
	}

	advertised := fmt.Sprintf("%v:%d", advertisedHost, advertisedPort)
	log.Printf("advertising as %+v\n", advertised)

	consulConfig := consulclient.DefaultConfig()
	consulConfig.Address = fmt.Sprintf("%v:%d", consulHost, consulPort)
	consul, err := consulclient.NewClient(consulConfig)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("connected to consul %+v\n", consul)
	agent := consul.Agent()

	// Register us as a service.  Each endpoint is listed as a tag.
	tags := make([]string, 0)
	for _, service := range config.Services {
		tags = append(tags, service.EndPoint)
	}

	service := consulclient.AgentServiceRegistration{
		ID:      "grunt-" + uuid.NewV4().String(),
		Name:    "grunt",
		Tags:    tags,
		Port:    advertisedPort,
		Address: advertisedHost,
	}
	err = agent.ServiceRegister(&service)
	if err != nil {
		log.Printf("error registering %v", err.Error())
	}

	log.Printf("Registered: %+v", service)

	// Register our check

	// We must check in with Consul every minute
	check := consulclient.AgentCheckRegistration{
		Name:      service.ID,
		ServiceID: service.ID,
	}
	check.TTL = "1m"

	err = agent.CheckRegister(&check)
	if err != nil {
		log.Printf("error registering check %v", err.Error())
	}

	ttl := func() {
		numberOfJobs := len(jobs)
		if numberOfJobs <= config.WarnLevel {
			agent.PassTTL(check.Name, fmt.Sprintf("%d jobs", numberOfJobs))
		} else if numberOfJobs <= config.CriticalLevel {
			agent.WarnTTL(check.Name, fmt.Sprintf("%d jobs", numberOfJobs))
		} else {
			agent.FailTTL(check.Name, fmt.Sprintf("%d jobs", numberOfJobs))
		}

	}

	go func() {
		ttl()
		// wake up every 30 seconds and check in
		ticker := time.NewTicker(time.Second * 30)
		for range ticker.C {
			ttl()
		}
	}()

	// when the program exits with SIGTERM and Interrupt (SIGINT), cleanly leave consul
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	signal.Notify(s, syscall.SIGTERM)
	go func() {
		<-s
		err := agent.ServiceDeregister(service.ID)
		if err != nil {
			log.Printf("Could not deregister %v\n", err.Error())
		}
		os.Exit(0)
	}()
}
