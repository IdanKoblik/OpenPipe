package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Rabbit RabbitConfig `yaml:"Rabbit"`
	Web WebConfig `yaml:"Web"`
}

type RabbitConfig struct {
	Channel string `yaml:"Channel"`
	Host string `yaml:"Host"`
	Username string `yaml:"Username"`
	Password string `yaml:"Password"`
	Port int `yaml:"Port"`	
}

type WebConfig struct {
	Host string `yaml:"Host"`
	Port int `yaml:"Port"`
}

func ReadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
   if err != nil {
		return nil, err
	}

	return &cfg, nil
}
