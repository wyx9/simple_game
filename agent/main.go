// agent/main.go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/yaml.v3"
)

type agentConfig struct {
	Listeners []listenerCfg `yaml:"Listeners"`
	Redis     redisCfg      `yaml:"Redis"`
}

type listenerCfg struct {
	Network string `yaml:"Network"`
	Addr    string `yaml:"Addr"`
	Port    string `yaml:"Port"`
}

type redisCfg struct {
	Addr     string `yaml:"Addr"`
	PassWord string `yaml:"PassWord"`
	DB       int    `yaml:"DB"`
}

func main() {
	data, err := os.ReadFile("agent/config/config.yaml")
	if err != nil {
		fmt.Println("load config failed:", err)
		os.Exit(1)
	}
	cfg := &agentConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		fmt.Println("parse config failed:", err)
		os.Exit(1)
	}

	startAgent(cfg.Listeners, cfg.Redis)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Agent server started")
	<-sigChan
	fmt.Println("Agent server shutting down...")
}
