package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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

type User struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func initDB(info *mgo.DialInfo) *mgo.Session {
	mongo, err := mgo.DialWithInfo(info)
	if err != nil {
		panic("db init failed")
	}
	return mongo
}

func jwtAuth(exclude []string) gin.HandlerFunc {
	return func(context *gin.Context) {

		for _, url := range exclude {
			if url == context.FullPath() {
				context.Next()
				return
			}
		}
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
			context.Abort()
			return
		}
		context.Set("claims", claims)
		context.Next()
	}
}

type JWT struct {
	SingingKey []byte
}

var (
	TokenExpired error = errors.New("token is expired")
	TokenMalformed error = errors.New("token is malformed")
	TokenInvalid error = errors.New("token is invalid")
	SignKey string = ""
)

type CustomChains struct {
	UID string
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

func (j *JWT) CreateToken(claims CustomChains) (string, error)  {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
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

func __main()  {

// ==============================================

	conf := Config{}

	var configFilepath string

	flag.StringVar(&configFilepath, "conf", "", "配置文件路径")

	flag.Parse()

	if configFilepath == "" {
		panic("请指定配置文件位置(.env.yaml)")
	}

	file, err := ioutil.ReadFile(configFilepath)

	if err != nil {
		panic(err)
	}

	configErr := yaml.Unmarshal(file, &conf)

	setSignKey(conf.Jwt.Secret)

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

// =================================================

	router := gin.Default()

	router.Use(jwtAuth([]string{"/auth", "/register"}))

	router.GET("/healthy", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": 1,
		})
	})
	
	router.POST("/register", func(ctx *gin.Context) {
		var user User
		if ctx.Bind(&user) == nil {
			client := mongo.DB("locationRecord").C("users")

			h:= sha1.New()
			h.Write([]byte(user.Password))

			data := model.User{
				Username: user.Username,
				Password: hex.EncodeToString(h.Sum(nil)),
				RegTime:  time.Now().Unix(),
			}
			err := client.Insert(data)
			if err != nil {
				log.Println(err)
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
				})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{
				"message": "register success",
			})
			return
		}
	})
	
	router.POST("/auth", func(ctx *gin.Context) {
		var user User
		if ctx.Bind(&user) == nil {
			client := mongo.DB("locationRecord").C("users")

			var userDoc model.User

			h := sha1.New()
			h.Write([]byte(user.Password))

			_ = client.Find(bson.M{
				"username": user.Username,
				"password": hex.EncodeToString(h.Sum(nil)),
			}).One(&userDoc)

			if userDoc.ID != "" {
				claims := CustomChains{
					UID:            userDoc.ID.Hex(),
				}
				j := NewJWT()
				token, err := j.CreateToken(claims)
				if err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					log.Println(err)
					return
				}

				ctx.JSON(http.StatusOK, gin.H{
					"data": gin.H{
						"token": token,
					},
				})
				return
			}
			ctx.JSON(http.StatusForbidden, gin.H{
				"message": "授权失败",
			})
			return
		}
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
		claims :=  ctx.MustGet("claims").(*CustomChains)
		var json Location
		if ctx.Bind(&json) == nil {
			fmt.Printf("lat %s, lng %s, city %s", json.Lat, json.Lng, json.City)
			client := mongo.DB("locationRecord").C("locations")

			data := model.Location{
				Lng:       json.Lng,
				Lat:       json.Lat,
				Timestamp: time.Now().Unix(),
				Uid:       bson.ObjectIdHex(claims.UID),
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