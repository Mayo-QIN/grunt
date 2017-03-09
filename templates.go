package main

import (
	"html/template"
	"net/http"

	"github.com/imdario/mergo"
)

func Template(name string, data map[string]interface{}, w http.ResponseWriter, request *http.Request) {
	var templateData = map[string]interface{}{
		"jobs":       jobs,
		"services":   config.Services,
		"serviceMap": config.ServiceMap,
	}
	// merge in our extra data
	mergo.Map(&templateData, data)
	contents, _ := Asset("template/" + name + ".html")
	t, _ := template.New(name).Parse(string(contents))
	t.Execute(w, templateData)
}
func Help(w http.ResponseWriter, request *http.Request) {
	Template("help", nil, w, request)
}
func Jobs(w http.ResponseWriter, request *http.Request) {
	Template("jobs", nil, w, request)
}
func Submit(w http.ResponseWriter, request *http.Request) {
	Template("submit", nil, w, request)
}
func Services(w http.ResponseWriter, request *http.Request) {
	Template("services", nil, w, request)
}
