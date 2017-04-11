package main

import (
	"encoding/xml"
	"fmt"
	"os/exec"
	"strings"
)

type Executable struct {
	Title            string           `xml:"title"`
	Description      string           `xml:"description"`
	Version          string           `xml:"version"`
	DocumentationURL string           `xml:"documentation-url"`
	License          string           `xml:"license"`
	Contributor      string           `xml:"contributor"`
	Acknowledgements string           `xml:"acknowledgements"`
	ParameterGroups  []ParameterGroup `xml:"parameters"`
}

type ParameterGroup struct {
	Parameters []Parameter `xml:",any"`
}

type Parameter struct {
	Name        string `xml:"name"`
	Longflag    string `xml:"longflag"`
	Flag        string `xml:"flag"`
	Description string `xml:"description"`
	Default     string `xml:"default"`
	Index       int    `xml:"index"`
	Channel     string `xml:"channel"`
}

// Run the executable at path, and return a service
func CreateService(path string) (*Service, error) {
	out, err := exec.Command(path, "--xml").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute %v: %v", path, err.Error())
	}

	return CreateServiceFromXML(path, string(out))
}

func CreateServiceFromXML(path string, out string) (*Service, error) {

	var executable Executable
	parser := xml.NewDecoder(strings.NewReader(out))
	err := parser.Decode(&executable)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %v", err.Error())
	}

	service := NewService()

	service.Description = executable.Description

	positional := make([]string, 20, 20)

	// Parse all the arguments
	for _, group := range executable.ParameterGroups {
		for _, p := range group.Parameters {

			// Not valid
			if p.Name == "" {
				continue
			}

			// Positional argument
			if p.Longflag == "" && p.Flag == "" {
				n := p.Name
				if p.Channel == "input" {
					n = "<" + n
				}
				if p.Channel == "output" {
					n = ">" + n
				}
				if p.Default != "" {
					service.Defaults[n] = p.Default
				}
				positional[p.Index] = n
			} else {
				// Flag
				n := "#" + p.Name
				if p.Channel == "output" {
					n = ">" + p.Name
				}
				if p.Channel == "input" {
					n = "<" + p.Name
				}
				f := ""
				if p.Longflag != "" {
					// Sometimes, the "--" is left on long flags...
					p.Longflag = strings.TrimPrefix(p.Longflag, "--")
					f = "--" + p.Longflag
				}
				if p.Flag != "" {
					// Sometimes, the "-" is left on flags...
					p.Flag = strings.TrimPrefix(p.Flag, "-")
					f = "-" + p.Flag
				}
				if p.Default != "" {
					service.Defaults[p.Name] = p.Default
				}
				if p.Description != "" {
					service.ParameterDescriptions[p.Name] = p.Description
				}
				service.CommandLine = append(service.CommandLine, f, n)
			}
		}
	}

	// Prepend the path, append the positional arguments
	service.CommandLine = append([]string{path}, service.CommandLine...)
	for _, arg := range positional {
		if arg != "" {
			service.CommandLine = append(service.CommandLine, arg)
		}
	}

	service.setup()
	return service, nil
}
