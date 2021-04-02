package main

import (
	"fmt"
	"io/ioutil"
	"rulecat/utils"

	log2 "rulecat/utils/log"

	"gopkg.in/yaml.v2"
)

type Input struct {
	Kafka struct {
		Enabled bool     `yaml:"enabled"`
		Server  []string `yaml:"server"`
		Topic   string   `yaml:"topic"`
		GroupId string   `yaml:"group_id"`
	}
}

type Output struct {
	Es struct {
		Enabled bool     `yaml:"enabled"`
		Server  []string `yaml:"es_host"`
	}
	Kafka struct {
		Enabled bool     `yaml:"enabled"`
		Server  []string `yaml:"server"`
		Topic   string   `yaml:"topic"`
		GroupId string   `yaml:"group_id"`
	}
	Json struct {
		Enabled bool   `yaml:"enabled"`
		Path    string `yaml:"path"`
	}
	Email struct {
		Enabled       bool   `yaml:"enabled"`
		EmailHost     string `yaml:"email_host"`
		EmailSmtpPort string `yaml:"email_smtp_port"`
		EmailFrom     string `yaml:"email_from"`
		EmailUserName string `yaml:"email_username"`
		EmailPwd      string `yaml:"email_pwd"`
	}
}

type Config struct {
	Name   string `yaml:"name"`
	Env    string `yaml:"env"`
	InPut  Input  `yaml:"input"`
	OutPut Output `yaml:"output"`
}

func main() {

	var err error
	configFile, err := ioutil.ReadFile(utils.GetCurrentPath() + "/etc/config.yml")
	if err != nil {
		log2.Error.Fatalf("Get yml file err %v ", err)

	}

	var _config *Config
	err = yaml.Unmarshal(configFile, &_config)

	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("config.app: %#v\n", _config)

	fmt.Println(_config.Name)
	fmt.Println(_config.Env)

	fmt.Println(_config.InPut)
	fmt.Println(_config.OutPut)

}
