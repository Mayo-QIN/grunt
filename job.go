package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	humanize "github.com/dustin/go-humanize"
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
	Address           []string          `json:"notification_email_address"`
	EndPoint          string            `json:"endpoint"`

	// Registered channels
	waiters []chan bool

	// Running process, NB: the Cmd is started in the Job's WorkingDirectory
	cmd    *exec.Cmd
	Output bytes.Buffer `json:"-"`
}

// Custom JSON output
// see http://choly.ca/post/go-json-marshalling/
func (job *Job) MarshalJSON() ([]byte, error) {
	type Alias Job
	f := func(t time.Time) string {
		if t.IsZero() {
			return ""
		} else {
			return humanize.Time(t)
		}
	}
	return json.Marshal(&struct {
		TempString         string `json:"output"`
		StartTimeHumanized string `json:"start_time_humanized"`
		EndTimeHumanized   string `json:"end_time_humanized"`
		*Alias
	}{
		TempString:         job.Output.String(),
		StartTimeHumanized: humanize.Time(job.StartTime),
		EndTimeHumanized:   f(job.EndTime),
		Alias:              (*Alias)(job),
	})
}

// Parse the commandline from the HTTP request
// The Job will run in the working directory, so paths must be relative.
// Store away enough information to get files back to the client.
func (job *Job) ParseCommandLine(request *http.Request) error {
	log.Printf("Parsing command line for %v", job.UUID)
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
			log.Printf("\tSaved uploaded file as %v (%v bytes)", header.Filename, count)
			fout.Close()
			cl = append(cl, filepath.Base(header.Filename))
		} else if prefix == '>' {
			// Write a file...
			if request.MultipartForm.Value[key] == nil {
				return fmt.Errorf("filename must be specified for %v", key)
			}
			job.FileMap[key] = filepath.Base(request.MultipartForm.Value[key][0])
			log.Printf("\tOutput file %v", job.FileMap[key])
			cl = append(cl, filepath.Base(request.MultipartForm.Value[key][0]))
		} else if prefix == '^' {
			// Expect a zip file, create a directory called key, unzip the contents and add to command line
			v := request.MultipartForm.File[key]
			if v == nil {
				return fmt.Errorf("Could not find %v in form data, was expecting uploaded zip file", key)
			}
			// Save the uploaded file
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
			fout.Close()
			defer os.Remove(zipFilename)

			// Make the directory, wrating the contents there
			dirName := filepath.Join(dir, filepath.Base(key))
			err = os.Mkdir(dirName, 0700)
			if err != nil {
				return err
			}
			err = Unzip(zipFilename, dirName)
			if err != nil {
				return err
			}

			// If the directory has a single entry and it's a directory, do a quick shuffle
			files, err := ioutil.ReadDir(dirName)
			if err != nil {
				return err
			}
			if len(files) == 1 && files[0].IsDir() {
				// Shuffle
				log.Printf("\tFound single directory, moving %v to %v", files[0].Name(), dirName)
				os.Rename(filepath.Join(dirName, files[0].Name()), dirName+"-temp")
				os.Remove(dirName)
				os.Rename(dirName+"-temp", dirName)
			}

			log.Printf("\tExtracted zip file to %v (%s)", key, humanize.Bytes(uint64(count)))
			cl = append(cl, filepath.Base(key))
		} else if prefix == '~' {
			if request.MultipartForm.Value[key] == nil {
				return fmt.Errorf("filename must be specified for %v", key)
			}

			job.ZipMap[key] = filepath.Base(request.MultipartForm.Value[key][0])
			if job.Service.CreateEmptyOutput {
				err := os.Mkdir(filepath.Join(dir, job.ZipMap[key]), 0700)
				if err != nil {
					return err
				}
			}
			log.Printf("\tOutput directory %v", job.ZipMap[key])
			cl = append(cl, job.ZipMap[key])
		} else {
			cl = append(cl, arg)
		}
	}
	log.Printf("Final command line: %v", cl)
	job.ParsedCommandLine = cl
	return nil
}

func (job *Job) Start() {
	cmd := exec.Command(job.ParsedCommandLine[0], job.ParsedCommandLine[1:]...)
	cmd.Dir = job.WorkingDirectory
	job.StartTime = time.Now()
	cmd.Stdout = &job.Output
	cmd.Stderr = &job.Output
	job.cmd = cmd
	job.Status = "pending"

	// Launch a go routine to wait
	go func() {
		jobMutex.Lock()
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
		jobMutex.Unlock()
		// Notify waiters
		log.Printf("%v (%v) completed with status %v", job.Service.EndPoint, job.UUID, job.Status)
		for _, c := range job.waiters {
			c <- true
		}

		// Send email here
		Email(job)

		// Cleanup after 120 minutes
		<-time.After(time.Minute * 120)
		Cleanup(job)
	}()

}

func (job *Job) Wait() {
	if job.Status == "running" {
		c := make(chan bool)
		job.Lock()
		job.waiters = append(job.waiters, c)
		job.Unlock()
		<-c
		close(c)
	}
}
