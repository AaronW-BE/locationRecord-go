package main

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

type Application struct {
	App *gin.Engine
	DB struct{
		MongoDB interface{}
	}
	Config Config
}

func (app *Application) Init(config Config)  {
	app.Config = config
	app.App = gin.Default()
}

func New(config Config) *Application  {
	application := &Application{}
	application.Init(config)
	return application
}

func (app *Application) Run()  {
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