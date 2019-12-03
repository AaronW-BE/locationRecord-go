package main

type Config struct {
	Application struct{
		Port int `yaml:"port"`
	}
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