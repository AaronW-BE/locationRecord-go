package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/yaml.v2"
	"github.com/dgrijalva/jwt-go"
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
	Mongodb struct{
		User string `yaml:"user"`
		Password string `yaml:"password"`
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	}
	Jwt struct{
		Secret string `yaml:"secret"`
		ExpireIn int `yaml:"expire_in"`
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

func jwtAuth() gin.HandlerFunc {
	return func(context *gin.Context) {
		token := context.Request.Header.Get("Authorization")
		if token == "" {
			context.JSON(http.StatusUnauthorized, gin.H{
				"status": 401,
				"msg":    "uncaught authorization token",
			})
			context.Abort()
			return
		}

		j := NewJWT()
		claims, err := j.parseToken(token)
		if err != nil {
			if err == TokenExpired {
				context.JSON(http.StatusUnauthorized, gin.H{
					"status": 401,
					"msg":    "token is expired",
				})
			}
			context.JSON(http.StatusUnauthorized, gin.H{
				"status": 401,
				"msg": err.Error(),
			})
		}
		context.Set("claims", claims)
		context.Next()
	}
}

type JWT struct {
	SingingKey []byte
}

var (
	TokenExpired error = errors.New("Token is expired")
	TokenMalformed error = errors.New("token is malformed")
	TokenInvalid error = errors.New("token is invalid")
	SignKey string = ""
)

type CustomChains struct {
	UID int `json:"uid"`
	jwt.StandardClaims
}

func NewJWT() *JWT {
	return &JWT{
		[]byte(getSignKey()),
	}
}

func getSignKey() string {
	return SignKey
}

func setSignKey(key string) string {
	SignKey = key
	return SignKey
}

func (j *JWT) CreateToken(clains CustomChains) (string, error)  {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, clains)
	return token.SignedString(j.SingingKey)
}

func (j *JWT) parseToken(tokenString string)(*CustomChains, error)  {
	token ,err := jwt.ParseWithClaims(tokenString, &CustomChains{}, func(token *jwt.Token) (i interface{}, err error) {
		return j.SingingKey, nil
	})

	if err != nil {
		if validationErr, ok := err.(*jwt.ValidationError); ok {
			if validationErr.Errors & jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			}else if validationErr.Errors & jwt.ValidationErrorExpired != 0 {
				return nil, TokenExpired
			} else {
				return nil, TokenInvalid
			}
		}
		
	}

	if claims, ok := token.Claims.(*CustomChains); ok && token.Valid {
		return claims, nil
	}
	return nil,TokenInvalid

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
		Addrs:          []string{conf.Mongodb.Host},
		Username:       conf.Mongodb.User,
		Password:       conf.Mongodb.Password,
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