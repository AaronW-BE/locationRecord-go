package main

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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
}

func New() *Application  {
	return &Application{}
}

func app()  {
	application := New()

	config := Config{}
	file, err := ioutil.ReadFile(".env.yaml")
	if err != nil {
		panic(err)
	}
	configErr := yaml.Unmarshal(file, &config)
	if configErr != nil{
		panic(configErr)
	}

	application.Init(config)

}