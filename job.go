package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type Job struct {
	sync.Mutex        `json:"-"`
	UUID              string            `json:"uuid"`
	CommandLine       []string          `yaml:"commandLine" json:"command_line"`
	Service           *Service          `json:"-"`
	ParsedCommandLine []string          `json:"-"`
	FileMap           map[string]string `json:"-"`
	ZipMap            map[string]string `json:"-"`
	WorkingDirectory  string            `json:"-"`
	StartTime         time.Time         `json:"start_time"`
	EndTime           time.Time         `json:"end_time"`
	Status            string            `json:"status"`
	Host              string            `json:"host"`
	Port              int               `json:"port"`
	Address           []string          `json:"address"`
	Endpoint          string            `json:"endpoint"`

	// Registered channels
	waiters []chan bool

	// Running process, NB: the Cmd is started in the Job's WorkingDirectory
	cmd    *exec.Cmd
	Output bytes.Buffer `json:"output"`
}

func (job *Job) ParseCommandLine(request *http.Request) error {
	cl := make([]string, 0)
	dir := job.WorkingDirectory
	for _, arg := range job.CommandLine {
		// Do we start with an #?
		key := arg[1:]
		prefix := arg[0]
		if prefix == '#' {
			// Lookup first in form
			if request.MultipartForm.Value[key] != nil {
				cl = append(cl, request.MultipartForm.Value[key][0])
			} else {
				// Look up in defaults
				cl = append(cl, job.Service.Defaults[key])
			}
		} else if prefix == '<' {
			// Do we have an < to indicate an uploaded file?
			v := request.MultipartForm.File[key]
			if v == nil {
				return fmt.Errorf("Could not find %v in form data", key)
			}
			header := v[0]
			// Save a temp file
			fout, err := os.Create(filepath.Join(dir, filepath.Base(header.Filename)))
			if err != nil {
				return err
			}
			f, err := header.Open()
			count, err := io.Copy(fout, f)
			if err != nil {
				return err
			}
			log.Printf("Wrote %v bytes to %v", count, header.Filename)
			fout.Close()
			cl = append(cl, fout.Name())
		} else if prefix == '>' {
			// Write a file...
			// Save a temp file
			if request.MultipartForm.Value[key] == nil {
				return fmt.Errorf("filename must be specified for %v", key)
			}
			tmp := filepath.Join(dir, filepath.Base(request.MultipartForm.Value[key][0]))
			job.FileMap[key] = tmp
			cl = append(cl, tmp)
		} else if prefix == '^' {
			// Expect a zip file, create a directory called key, unzip the contents and add to command line
			v := request.MultipartForm.File[key]
			if v == nil {
				return fmt.Errorf("Could not find %v in form data, was expecting uploaded zip file", key)
			}
			header := v[0]
			f, err := header.Open()
			zipFilename := filepath.Join(dir, filepath.Base(header.Filename)+".zip")
			fout, err := os.Create(zipFilename)
			if err != nil {
				return err
			}
			count, err := io.Copy(fout, f)
			if err != nil {
				return err
			}
			log.Printf("Wrote %v bytes to %v", count, header.Filename)
			fout.Close()

			// Make the directory
			dirName := filepath.Join(dir, filepath.Base(key))
			err = os.Mkdir(dirName, 0700)
			if err != nil {
				return err
			}
			err = Unzip(zipFilename, dirName)
			if err != nil {
				return err
			}
			cl = append(cl, dirName)
		} else if prefix == '~' {
			if request.MultipartForm.Value[key] == nil {
				return fmt.Errorf("filename must be specified for %v", key)
			}
			dirName := filepath.Join(dir, filepath.Base(key))
			if job.Service.CreateEmptyOutput {
				err := os.Mkdir(dirName, 0700)
				if err != nil {
					return err
				}
			}
			job.ZipMap[key] = dirName
			cl = append(cl, dirName)
		} else {
			cl = append(cl, arg)
		}
	}
	log.Printf("Final command line: %v", cl)
	job.ParsedCommandLine = cl
	return nil
}
