package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type ConfigD struct {
	Name     string
	Services []*Service
}

func loadServices(configDirectory string) error {
	filepath.Walk(configDirectory, func(gruntfile string, info os.FileInfo, err error) error {
		// Don't do anything with directories
		if info.IsDir() {
			return nil
		}

		// We simply abort on any error
		if err != nil {
			return err
		}

		log.Printf("loading config file %v\n", gruntfile)
		// read the file, and add the services to the global config
		data, err := ioutil.ReadFile(gruntfile)
		if err != nil {
			return fmt.Errorf("Error reading %v: %v ", gruntfile, err)
		}
		var configD ConfigD
		err = yaml.Unmarshal(data, &configD)
		if err != nil {
			return fmt.Errorf("Error in YML parsing: %v", err)
		}

		// Advertise in Consul
		registerConfigWithConsul(&configD)

		// Append to existing service endpoints
		config.Services = append(config.Services, configD.Services...)
		return nil
	})
	return nil
}
