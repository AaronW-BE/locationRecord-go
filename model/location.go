package model

import "gopkg.in/mgo.v2/bson"

type Location struct {
	Lat string
	Lng string
	Timestamp int64
	Uid bson.ObjectId
}

type User struct {
	ID bson.ObjectId `json:"Id,omitempty" bson:"_id,omitempty"`
	Username string
	Password string
	RegTime int64
}