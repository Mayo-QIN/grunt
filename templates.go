package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"

	"github.com/Mayo-QIN/grunt/dassets"
	"github.com/gorilla/mux"
	"github.com/imdario/mergo"
	"github.com/russross/blackfriday"
)

// Add some helper functions
var funcs = template.FuncMap{
	"json": func(v interface{}) (string, error) {
		a, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(a), nil
	},
	"humanizeTime": humanize.Time,
	"now":          time.Now,
	"isArray":      func(s string) bool { return strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") },
	"toArray": func(s string) []string {
		v := strings.TrimPrefix(s, "[")
		v = strings.TrimSuffix(v, "]")
		return strings.Split(v, ",")
	},
	"markdown": func(s string) template.HTML {
		return template.HTML(string(blackfriday.MarkdownCommon([]byte(s))))
	},
}

var templates = template.New("").Funcs(funcs)
var helpText = template.HTML("")

func init() {
	loadTemplates(AssetNames(), Asset)
}

func loadTemplates(names []string, asset func(name string) ([]byte, error)) {
	for _, path := range names {
		bytes, err := asset(path)
		if err != nil {
			if debug {
				log.Printf("Unable to parse: path=%s, err=%s", path, err)
			} else {
				log.Panicf("Unable to parse: path=%s, err=%s", path, err)
			}
		}
		templates.New(path).Parse(string(bytes))
	}
	bytes, _ := asset("README.md/README.md")
	helpText = template.HTML(string(blackfriday.MarkdownCommon(bytes)))
}

func Template(name string, data map[string]interface{}, w http.ResponseWriter, request *http.Request) {
	templateData := map[string]interface{}{
		"jobs":           jobs,
		"config":         config,
		"services":       config.Services,
		"serviceMap":     config.ServiceMap,
		"help":           helpText,
		"vars":           mux.Vars(request),
		"version":        Version,
		"fullVersion":    FullVersion,
		"buildTimestamp": BuildTimestamp,
		"hash":           Hash,
	}

	// merge in our extra data
	mergo.Map(&templateData, data)
	if debug {
		templates = template.New("").Funcs(funcs)
		loadTemplates(dassets.AssetNames(), dassets.Asset)
	}

	if templates.Lookup("template/"+name) == nil {
		http.Error(w, name+" is not found", http.StatusNotFound)
		return
	}

	err := templates.ExecuteTemplate(w, "template/"+name, templateData)
	if err != nil {
		log.Printf("error in template %v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ExecTemplate(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	name := vars["name"]
	Template(name, nil, w, request)
}

// Return the information about a job as JSON
func JobDetail(w http.ResponseWriter, request *http.Request) {
	key := mux.Vars(request)["id"]
	job := jobs[key]
	if job == nil {
		http.Error(w, "could not find job", http.StatusNotFound)
		return
	}
	var data = map[string]interface{}{"job": job}
	Template("job.html", data, w, request)
}

// Return the information about a job as JSON
func ServiceDetail(w http.ResponseWriter, request *http.Request) {
	key := mux.Vars(request)["id"]
	service, ok := config.ServiceMap[key]
	if !ok {
		http.Error(w, fmt.Sprintf("could not find service %v", key), http.StatusNotFound)
		return
	}
	var data = map[string]interface{}{"service": service}
	Template("service.html", data, w, request)
}
