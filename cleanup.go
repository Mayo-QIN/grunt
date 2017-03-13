package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"strings"
	"text/template"
)

func Email(job *Job) {
	if config.Mail.Server == "" || len(job.Address) == 0 {
		return
	}
	var auth smtp.Auth = nil
	if config.Mail.Username != "" && config.Mail.Password != "" {
		auth = smtp.PlainAuth("", config.Mail.Username, config.Mail.Password, config.Mail.Server)
	}

	var templateData = map[string]interface{}{
		"job":     job,
		"service": config.ServiceMap[job.EndPoint],
		"to":      strings.Trim(fmt.Sprint(job.Address), "[]"),
		"config":  config,
	}
	contents, _ := Asset("template/email.txt")
	t, _ := template.New("email").Parse(string(contents))
	var buffer bytes.Buffer
	err := t.Execute(&buffer, templateData)
	if err != nil {
		log.Printf("Failed to construct email: %v", err.Error())
		return
	}
	server := fmt.Sprintf("%s:%d", config.Mail.Server, config.Mail.Port)
	log.Printf("Mailing status for %v", job.UUID)
	log.Printf("Sending to: %v", templateData["job"])
	log.Printf("With auth: %+v", auth)
	log.Printf("Body: \n%v", buffer.String())
	err = smtp.SendMail(server, auth, config.Mail.From, job.Address, buffer.Bytes())
	if err != nil {
		log.Printf("Error sending mail: %v", err.Error())
	} else {
		log.Printf("Mail sent")
	}
}

func Cleanup(job *Job) {
	log.Printf("Cleaning up %v", job.UUID)

	dir, err := ioutil.TempDir("", job.UUID)
	if err != nil {
		log.Printf("Error creating temp dir path: %v", err.Error())
		return
	}
	log.Printf("Deleting temp directory %v", dir)
	err = os.RemoveAll(dir)
	if err != nil {
		log.Printf("Error removing dir: %v", err.Error())
		return
	}
	delete(jobs, job.UUID)
	log.Printf("Cleanup done of %v (%v)", job.EndPoint, job.UUID)
}
