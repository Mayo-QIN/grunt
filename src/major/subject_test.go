package main

import (
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

func TestSubject(t *testing.T) {
	subject := Subject{}
	tt := reflect.New(reflect.TypeOf(subject)).Interface()
	fmt.Printf("Type is %v\n", reflect.TypeOf(subject))
	fmt.Printf("Object is: %v\n", tt)

	reflect.ValueOf(tt).Elem().FieldByName("Name").SetString("hi")
	reflect.ValueOf(tt).Elem().FieldByName("Id").Set(reflect.ValueOf(bson.NewObjectId()))
	fmt.Printf("Object is: %v\n", tt)

	// fmt.Printf("Object is settable: %v\n", tt.CanSet())
	// fmt.Printf("ID is: %v\n", tt.FieldByName("Id"))
	// fmt.Printf("ID is settable?: %v\n", tt.FieldByName("Id").CanSet())
	// tt.FieldByName("Id").Set(reflect.ValueOf(bson.NewObjectId()))

	// fmt.Printf("Id is: %v\n", reflect.Indirect(tt).FieldByName("Id"))

}
