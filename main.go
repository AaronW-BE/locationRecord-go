package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"location-record/middleware"
	"location-record/model"
	"log"
	"net/http"
	"time"
)

type Location struct {
	Lat string `form:"lat" json:"lat" binding:"required"`
	Lng string `form:"lng" json:"lng" binding:"required"`
	City string `form:"city" json:"city" binding:"required"`
}


type Config struct {
	MONGODB struct{
		User string `yaml:"user"`
		Password string `yaml:"password"`
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	}
}

func initDB(info *mgo.DialInfo) *mgo.Session {
	mongo, err := mgo.DialWithInfo(info)
	if err != nil {
		log.Fatalf("create mongo session failed %s \n", err)
		panic("db init failed")
	}
	return mongo
}

func main()  {
	conf := Config{}

	file, err := ioutil.ReadFile(".env.yaml")

	if err != nil {
		panic(err)
	}

	configErr := yaml.Unmarshal(file, &conf)

	if configErr != nil{
		panic(configErr)
	}

	dialInfo := &mgo.DialInfo{
		Addrs:          []string{conf.MONGODB.Host},
		Username:       conf.MONGODB.User,
		Password:       conf.MONGODB.Password,
	}

	mongo :=initDB(dialInfo)

	defer mongo.Close()

	router := gin.Default()

	router.Use(middleware.Auth([]string{"/"}))

	router.GET("/healthy", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": 1,
		})
	})

	router.GET("/location", func(ctx *gin.Context) {
		client := mongo.DB("locationRecord").C("locations")

		var locationSet []model.Location

		currentDayTmp := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0,0,0,0, time.Local).Unix()
		_ = client.Find(bson.M{"timestamp": bson.M{
			"$gt":  currentDayTmp,
			"$lte": time.Now().Unix(),
		}}).All(&locationSet)

		ctx.JSON(http.StatusOK, gin.H{"data": locationSet, "message": "ok",})
	})

	router.PUT("/location", func(ctx *gin.Context) {
		var json Location
		if ctx.Bind(&json) == nil {
			fmt.Printf("lat %s, lng %s, city %s", json.Lat, json.Lng, json.City)
			client := mongo.DB("locationRecord").C("locations")

			data := model.Location{
				Lng:       json.Lng,
				Lat:       json.Lat,
				Timestamp: time.Now().Unix(),
			}
			err := client.Insert(data)
			if err != nil {
				log.Fatal(err)
			}
			ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
		}
	})

	_ = http.ListenAndServe(":8000", router)

}