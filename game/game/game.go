package game

import (
	"fmt"
	"os"
	"os/signal"
	"simple_game/game/config"
	"simple_game/game/http"
	"simple_game/game/pkg"
	"simple_game/game/register"
	"simple_game/game/routes"
	"simple_game/game/server"
	"syscall"
)

var printStart = "\n     _                 _                                   \n ___(_)_ __ ___  _ __ | | ___    __ _  __ _ _ __ ___   ___ \n/ __| | '_ ` _ \\| '_ \\| |/ _ \\  / _` |/ _` | '_ ` _ \\ / _ \\\n\\__ \\ | | | | | | |_) | |  __/ | (_| | (_| | | | | | |  __/\n|___/_|_| |_| |_| .__/|_|\\___|  \\__, |\\__,_|_| |_| |_|\\___|\n                |_|             |___/                      \n"

func Start(params ...string) {
	// 路由注册
	fmt.Print(printStart)
	// 日志
	pkg.StartLog()
	// 基础配置
	config.Start()

	// 信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 路由注册
	routes.Init()
	register.RegisteredRoute()

	// 核心服务启动
	pkg.MysqlStart()
	pkg.RedisStart()

	go server.StartGRPC()
	go http.StartHttp()
	go server.Start()

	<-sigChan
	Stop()
}

func Stop() {
	fmt.Println("simple game stop  start save")
	// todo 暂时屏蔽
	// todo 改成用chan
	//http.StartHttp()
	//err := server.StopServer()
	//if err != nil {
	//	return
	//}
}
