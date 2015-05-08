package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

type Subject struct {
	Id   bson.ObjectId     `bson:"_id,omitempty" json:"id"`
	Name string            `json:"name"`
	Info map[string]string `json:"info"`
}

func registerSubject(r *mux.Router) {
	r.Path("/{collection}/search").Methods("POST").HandlerFunc(searchSubject)
	r.Path("//{collection}/{id}").Methods("GET").HandlerFunc(getSubject)
	r.Path("/{collection}/").Methods("POST").HandlerFunc(createSubject)
	r.Path("/{collection}/{id}").Methods("PUT").HandlerFunc(updateSubject)
	r.Path("/{collection}/{id}").Methods("DELETE").HandlerFunc(deleteSubject)
}

func createSubject(w http.ResponseWriter, request *http.Request) {
	var subject map[string]interface{}
	err := json.NewDecoder(request.Body).Decode(&subject)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	subject["_id"] = bson.NewObjectId()
	c := session.DB(database).C(mux.Vars(request)["collection"])
	err = c.Insert(&subject)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(subject)

}

func getSubject(w http.ResponseWriter, request *http.Request) {
	c := session.DB(database).C(mux.Vars(request)["collection"])
	key := mux.Vars(request)["id"]
	result := bson.M{}
	err := c.FindId(bson.ObjectIdHex(key)).One(&result)
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
	c := session.DB(database).C(mux.Vars(request)["collection"])
	var subject map[string]interface{}
	err := json.NewDecoder(request.Body).Decode(&subject)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	key := mux.Vars(request)["id"]
	err = c.UpdateId(bson.ObjectIdHex(key), &subject)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(subject)
}

func searchSubject(w http.ResponseWriter, request *http.Request) {
	c := session.DB(database).C(mux.Vars(request)["collection"])
	buffer := make([]byte, request.ContentLength)
	request.Body.Read(buffer)
	var query bson.M
	bson.Unmarshal(buffer, &query)
	var result []interface{}
	err := c.Find(query).All(&result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(bson.M{mux.Vars(request)["collection"]: result})
}
