package main

import (
	"encoding/json"
	"log"
	"os"
)

type Account struct {
	Filter            string `json:"filter"`
	SkipTLSValidation bool   `json:"skip_tls_validation"`
	Pem               string `json:"pem"`
	URL               string `json:"url"`
}

type Config struct {
	Listen    string    `json:"listen"`
	TimeoutMs int       `json:"timeout"`
	Accounts  []Account `json:"accounts"`
}

func NewConfig(filename string) *Config {
	fd, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer fd.Close()

	config := &Config{
		Listen: ":9091",
	}
	if err := json.NewDecoder(fd).Decode(config); err != nil {
		log.Fatalln(err)
	}

	return config
}
