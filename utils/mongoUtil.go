package utils

import (
	"gopkg.in/mgo.v2"
)

func InitMongo(info *mgo.DialInfo) *mgo.Session {
	mongo, err := mgo.DialWithInfo(info)
	if err != nil {
		panic("db init failed")
	}
	defer mongo.Close()
	return mongo
}