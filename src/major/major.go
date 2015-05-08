package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	"github.com/codegangsta/cli"
	mgo "gopkg.in/mgo.v2"
	graceful "gopkg.in/tylerb/graceful.v1"
)

var session *mgo.Session
var database string

func RunMajor(c *cli.Context) {
	var err error
	log.Info("Hi from major")
	log.WithFields(log.Fields{
		"host":     c.String("host"),
		"port":     c.Int("port"),
		"database": c.String("database"),
	}).Info("Connecting to MongoDB")

	database = c.String("database")
	url := fmt.Sprintf("%v:%v", c.String("host"), c.Int("port"))
	session, err = mgo.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect with Mongo on %v: %v", url, err)
	}

	log.Infof("Connected to MongoDB @ %v", url)

	// Start the service
	r := mux.NewRouter()
	s := r.PathPrefix("/rest").Subrouter()
	registerSubject(s)
	http.Handle("/rest/", r)

	log.Info("Starting major on port 9902")
	hn, err := os.Hostname()
	log.Infof("http://%v:9902", hn)
	addresses, err := net.LookupHost(hn)
	// handle err
	for _, addr := range addresses {
		log.Infof("http://%v:9902", addr)
	}
	graceful.Run(":9902", 10*time.Second, nil)

}
