package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
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

	// 尝试多个可能的配置文件路径
	possiblePaths := []string{
		"./config/config.yaml",           // 相对于当前目录
		pwd + "/config/config.yaml",      // 绝对路径
		pwd + "/game/config/config.yaml", // game 子目录
		"../config/config.yaml",          // 上级目录
	}

	var configFile string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			configFile = path
			break
		}
	}

	if configFile == "" {
		fmt.Println("无法找到配置文件，请确保 config.yaml 存在于以下路径之一：")
		for _, path := range possiblePaths {
			fmt.Println(" - " + path)
		}
		return
	}

	bytes, err := ioutil.ReadFile(configFile)
	if err != nil || len(bytes) == 0 {
		fmt.Println("读取配置文件失败:", err)
		return
	}

	if err := yaml.Unmarshal(bytes, &Conf); err != nil {
		fmt.Println("解析配置文件失败:", err)
		return
	}

}
