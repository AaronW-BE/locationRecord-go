package main

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"location-record/model"
	"location-record/types"
	"location-record/utils"
	"net/http"
	"time"
)

type Application struct {
	Engine *gin.Engine
	DB struct{
		MongoDB *mgo.Session
	}
	Config types.Config
}

func (app *Application) Init(config types.Config)  {
	app.Config = config
	app.Engine = gin.Default()

	// TODO process middleware

	// process initial db
	app.DB.MongoDB = utils.InitMongo(config)

	// TODO process initial router
	app.Engine.GET("/", func(context *gin.Context) {
		client := app.DB.MongoDB.DB("locationRecord").C("users")
					data := model.User{
						Username: "test",
						Password: "",
						RegTime:  time.Now().Unix(),
					}
					err := client.Insert(data)

		context.JSON(200, gin.H{
			"STATUS": 1,
			"err":    err,
		})
	})
}

func New(config types.Config) *Application  {
	application := &Application{}
	application.Init(config)
	return application
}

func (app *Application) Run()  {
	_ = http.ListenAndServe(":8000", app.Engine)
}

func (app *Application) RegisterRoute()  {

}

func main()  {
	config := types.Config{}
	file, err := ioutil.ReadFile(".env.yaml")
	if err != nil {
		panic(err)
	}
	configErr := yaml.Unmarshal(file, &config)
	if configErr != nil{
		panic(configErr)
	}
	application := New(config)
	defer application.DB.MongoDB.Close()
	application.Run()
}