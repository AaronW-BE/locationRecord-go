package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"location-record/types"
)

func main() {
	config := types.Config{}
	file, err := ioutil.ReadFile(".env.yaml")
	if err != nil {
		panic(err)
	}
	configErr := yaml.Unmarshal(file, &config)
	if configErr != nil {
		panic(configErr)
	}
	application := New(config)
	defer application.DB.MongoDB.Close()
	application.Run()
}
