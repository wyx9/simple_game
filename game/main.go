// game/main.go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	engine "simple_game/game/engine"
	"simple_game/game/pkg"
	"simple_game/game/register"
	"simple_game/game/routes"
)

func main() {
	// 加载配置
	cfg, err := engine.LoadGameConfig("game/config/config.yaml")
	if err != nil {
		fmt.Println("load config failed:", err)
		os.Exit(1)
	}

	// 日志
	pkg.StartLog()
	if cfg.SaveLog {
		_ = pkg.InitLogFile()
	}

	// 路由注册
	routes.Init()
	register.RegisteredRoute()

	// 基础设施
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cfg.MySql.User, cfg.MySql.PassWord, cfg.MySql.Addr, cfg.MySql.Port, cfg.MySql.DBName)
	pkg.MysqlStart(dsn)
	pkg.RedisStart(cfg.Redis.Addr, cfg.Redis.PassWord, cfg.Redis.DB)

	// 信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动服务
	go engine.StartGRPC()
	go engine.Start(
		fmt.Sprintf("%s:%s", cfg.Listen.Addr, cfg.Listen.Port),
		cfg.TokenSecret,
	)

	fmt.Println("Game server started")
	<-sigChan
	fmt.Println("Game server shutting down...")
}
