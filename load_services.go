package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type ConfigD struct {
	Name           string
	Services       []*Service
	SlicerServices []*SlicerService `yml:"cli"`
}

func loadServices(configDirectory string) error {
	return filepath.Walk(configDirectory, func(gruntfile string, info os.FileInfo, err error) error {
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

		for _, ss := range configD.SlicerServices {
			s, err := CreateService(ss.Executable)
			if err != nil {
				return fmt.Errorf("Error constructing Slicer CLI: %v", err)
			}
			config.Services = append(config.Services, s)
		}
		for _, s := range configD.Services {
			config.Services = append(config.Services, s)
		}
		return nil
	})
}
