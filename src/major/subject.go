package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"

	mgo "gopkg.in/mgo.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

type Subject struct {
	Id   bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name string        `json:"name"`
}

type SubjectInfo struct {
	Id        bson.ObjectId          `bson:"_id,omitempty" json:"id"`
	SubjectId bson.ObjectId          `bson:"subject_id,omitempty" json:"subject_id"`
	Tags      map[string]interface{} `bson:"tags" json:"tags"`
}

type Study struct {
	Id               bson.ObjectId `bson:"_id,omitempty" json:"id"`
	SubjectId        bson.ObjectId `bson:"subject_id,omitempty" json:"subject_id"`
	StudyDescription string        `bson:"study_description" json:"study_description"`
	StudyDate        time.Time     `bson:"study_date" json:"study_date"`
}

type StudyInfo struct {
	Id      bson.ObjectId          `bson:"_id,omitempty" json:"id"`
	StudyId bson.ObjectId          `bson:"study_id,omitempty" json:"study_id"`
	Tags    map[string]interface{} `bson:"tags" json:"tags"`
}

type Series struct {
	Id                bson.ObjectId `bson:"_id,omitempty" json:"id"`
	StudyId           bson.ObjectId `bson:"study_id" json:"study_id"`
	SeriesDescription string        `bson:"series_description" json:"series_description"`
}

type SeriesInfo struct {
	Id       bson.ObjectId          `bson:"_id,omitempty" json:"id"`
	SeriesId bson.ObjectId          `bson:"series_id,omitempty" json:"series_id"`
	Tags     map[string]interface{} `bson:"tags" json:"tags"`
}

type Snapshot struct {
	Id       bson.ObjectId          `bson:"_id,omitempty" json:"id"`
	SeriesId bson.ObjectId          `bson:"series_id,omitempty" json:"series_id"`
	Name     string                 `bson:"name" json:"name"`
	Tags     map[string]interface{} `bson:"tags" json:"tags"`
	FileId   bson.ObjectId          `bson:"file_id,omitempty" json:"file_id"`
}

var TypeMap = map[string]reflect.Type{
	"subject":      reflect.TypeOf(Subject{}),
	"subject_info": reflect.TypeOf(SubjectInfo{}),
	"study":        reflect.TypeOf(Study{}),
	"study_info":   reflect.TypeOf(StudyInfo{}),
	"series":       reflect.TypeOf(Series{}),
	"series_info":  reflect.TypeOf(SeriesInfo{}),
	"snapshot":     reflect.TypeOf(Snapshot{}),
}

func registerSubject(r *mux.Router) {

	r.Path("/rest/{collection}/test").HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(w, "test -- %v\n", mux.Vars(request)["collection"])
	})

	r.Path("/rest/snapshot/{id}/file").Methods("POST", "PUT").HandlerFunc(uploadFile)
	// // r.Path("/rest/snapshot/{id}/file").Methods("GET").HandlerFunc(downloadFile)

	r.Path("/rest/{collection}/search").Methods("POST").HandlerFunc(searchSubject)
	r.Path("/rest/{collection}/{id}").Methods("GET").HandlerFunc(getSubject)
	r.Path("/rest/{collection}").Methods("POST").HandlerFunc(createSubject)
	r.Path("/rest/{collection}/{id}").Methods("PUT").HandlerFunc(updateSubject)
	r.Path("/rest/{collection}/{id}").Methods("DELETE").HandlerFunc(deleteSubject)

	log.Info("Registered Subject with the router")
}

func createSubject(w http.ResponseWriter, request *http.Request) {
	log.Info("Creating object")
	collection := mux.Vars(request)["collection"]
	t, ok := TypeMap[collection]
	if !ok {
		http.Error(w, "unknown type", http.StatusInternalServerError)
		return
	}
	log.Infof("Creating object for collection %v", collection)
	subject := reflect.New(t).Interface()
	err := json.NewDecoder(request.Body).Decode(subject)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the Id field
	reflect.ValueOf(subject).Elem().FieldByName("Id").Set(reflect.ValueOf(bson.NewObjectId()))
	c := session.DB(database).C(collection)
	log.Info("About to insert")
	err = c.Insert(subject)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(subject)

}

func getObject(collection, key string) (interface{}, error) {
	c := session.DB(database).C(collection)
	t, ok := TypeMap[collection]
	if !ok {
		return nil, errors.New("Could not find proper type")
	}
	result := reflect.New(t).Interface()
	log.Infof("Creating object %v for collection %v of type %v", result, collection, t)
	err := c.FindId(bson.ObjectIdHex(key)).One(result)
	log.Infof("Created object %v of type %v", result, reflect.TypeOf(result))

	return result, err
}

func getSubject(w http.ResponseWriter, request *http.Request) {
	log.Info("Get")
	collection := mux.Vars(request)["collection"]
	key := mux.Vars(request)["id"]
	result, err := getObject(collection, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func deleteSubject(w http.ResponseWriter, request *http.Request) {
	c := session.DB(database).C(mux.Vars(request)["collection"])
	key := mux.Vars(request)["id"]
	err := c.RemoveId(bson.ObjectIdHex(key))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func updateSubject(w http.ResponseWriter, request *http.Request) {
	collection := mux.Vars(request)["collection"]
	c := session.DB(database).C(collection)

	log.Infof("Updating object for collection %v", collection)
	t, ok := TypeMap[collection]
	if !ok {
		http.Error(w, "unknown type", http.StatusInternalServerError)
		return
	}

	subject := reflect.New(t).Interface()
	err := json.NewDecoder(request.Body).Decode(subject)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	key := mux.Vars(request)["id"]
	err = c.UpdateId(bson.ObjectIdHex(key), subject)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	result, err := getObject(collection, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func searchSubject(w http.ResponseWriter, request *http.Request) {
	collection := mux.Vars(request)["collection"]
	c := session.DB(database).C(collection)

	log.Infof("Updating object for collection %v", collection)
	t, ok := TypeMap[collection]
	if !ok {
		http.Error(w, "unknown type", http.StatusInternalServerError)
		return
	}

	var query bson.M
	err := json.NewDecoder(request.Body).Decode(&query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Infof("Query %v", query)

	result := make([]interface{}, 0)

	iter := c.Find(query).Iter()
	subject := reflect.New(t).Interface()
	for iter.Next(subject) {
		result = append(result, subject)
		subject = reflect.New(t).Interface()
	}

	json.NewEncoder(w).Encode(bson.M{mux.Vars(request)["collection"]: result})
}

func uploadFile(w http.ResponseWriter, request *http.Request) {
	log.Info("uploadFile")
	collection := "snapshot"
	c := session.DB(database).C(collection)
	gridFS := session.DB(database).GridFS("fs")
	key := mux.Vars(request)["id"]
	sn, err := getObject(collection, key)
	log.Infof("Object: %v / %v", sn, reflect.TypeOf(sn))
	if snapshot, found := sn.(*Snapshot); found {
		log.Infof("Uploading a file to %v / %v", snapshot.FileId, snapshot.FileId.Hex())
		// Upload a file
		var gridFile *mgo.GridFile
		if snapshot.FileId.Hex() == "" {
			log.Info("FileId not valid, creating a new file")
			gridFile, err = gridFS.Create("")
			snapshot.FileId = gridFile.Id().(bson.ObjectId)
			// Save the snapshot
			c.UpdateId(snapshot.Id, snapshot)
		} else {
			log.Info("FileId valid, opening file")
			gridFile, err = gridFS.OpenId(snapshot.FileId)
		}
		log.Info("Writing file")
		_, err = io.Copy(gridFile, request.Body)
		gridFile.Close()
		request.Body.Close()
		json.NewEncoder(w).Encode(snapshot)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
