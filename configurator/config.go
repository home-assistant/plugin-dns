package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type CoreDNSConfig struct {
	Servers []string `json:"servers"`
	Locals  []string `json:"locals"`
	Debug   bool     `json:"debug"`
}

func readConfigFile(file string) *CoreDNSConfig {
	configFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	config := CoreDNSConfig{}
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatal(err)
	}

	return &config
}
