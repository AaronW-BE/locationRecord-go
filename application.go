package main

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

type Application struct {
	App *gin.Engine
	DB struct{
		MongoDB *mgo.Session
	}
	Config Config
}

func (app *Application) Init(config Config)  {
	app.Config = config
	app.App = gin.Default()

	// TODO process middleware

	// TODO process initial db
	//dialInfo := &mgo.DialInfo{
	//	Addrs:          []string{config.Mongodb.Host},
	//	Username:       config.Mongodb.User,
	//	Password:       config.Mongodb.Password,
	//}
	//app.DB.MongoDB = utils.InitMongo(dialInfo)

	// TODO process initial router
}

func New(config Config) *Application  {
	application := &Application{}
	application.Init(config)
	return application
}

func (app *Application) Run()  {
	app.App.GET("/", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"STATUS": 1,
		})
	})
	println("running")
	_ = http.ListenAndServe(":8000", app.App)
}

func main()  {
	config := Config{}
	file, err := ioutil.ReadFile(".env.yaml")
	if err != nil {
		panic(err)
	}
	configErr := yaml.Unmarshal(file, &config)
	if configErr != nil{
		panic(configErr)
	}
	application := New(config)
	application.Run()
}