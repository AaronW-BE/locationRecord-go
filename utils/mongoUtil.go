package utils

import (
	"gopkg.in/mgo.v2"
	"location-record/types"
	"time"
)

func InitMongo(config types.Config) *mgo.Session {
	dialInfo := &mgo.DialInfo{
		Addrs:          []string{config.Mongodb.Host},
		Username:       config.Mongodb.User,
		Password:       config.Mongodb.Password,
		Timeout:		time.Second * time.Duration(config.Mongodb.Timeout),
	}
	mongo, err := mgo.DialWithInfo(dialInfo)

	if err != nil {
		panic("db init failed")
	}
	return mongo
}