package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

var configFilePath = "./config/config.yaml"

var configFilePath2 = "/app/config/config.yaml"

// Note: struct fields must be public in order for unmarshal to
// correctly populate the data.
type Config struct {
	OpenGm bool `yaml:"OpenGm"`
	MySql  struct {
		Addr     string `yaml:"Addr"`
		Port     int    `yaml:"Port"`
		User     string `yaml:"User"`
		PassWord string `yaml:"PassWord"`
		DBName   string `yaml:"DBName"`
	} `yaml:"MySql"`

	Redis struct {
		Addr     string `yaml:"Addr"`
		PassWord string `yaml:"PassWord"`
		DB       int    `yaml:"DB"`
	} `yaml:"Redis"`

	Tcp struct {
		Addr string `yaml:"Addr"`
		Port string `yaml:"Port"`
	} `yaml:"Tcp"`

	Http struct {
		Addr string `yaml:"Addr"`
		Port string `yaml:"Port"`
	} `yaml:"Http"`

	Rpc struct {
		Addr string `yaml:"Addr"`
		Port string `yaml:"Port"`
	} `yaml:"Rpc"`
}

var Conf = Config{}

func Start() {
	pwd, _ := os.Getwd()
	cfgFail := pwd + configFilePath2
	bytes, err := ioutil.ReadFile(cfgFail)
	if err != nil || len(bytes) == 0 {
		fmt.Println("conf yaml init error:", err)
	}
	if err := yaml.Unmarshal(bytes, &Conf); err != nil {
		fmt.Println(Conf.OpenGm)
	}
}
