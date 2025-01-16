package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func Test_Start(t *testing.T) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	logger.Start()
	// 路由注册
	fmt.Print(printStart)
	register.RegisteredRoute()

	utils.MysqlStart()
	utils.RedisStart()
	go server.StartGRPC()
	go http.StartHttp()
	gs := server.NewServer("127.0.0.1", 8888)
	go gs.Start()
	// 初始化actor
	actor.Init()

	go func() {
		for {
			time.Sleep(time.Second * 10)
			test()
		}
	}()
	<-sigChan
	Stop()
}

func test() {
	pidList := make([]int64, 0)
	actor.Directors.ActorMap.Range(func(key, value any) bool {
		iActor := value.(actor.IActor)
		if player, ok := iActor.(server.Player); ok {
			pidList = append(pidList, player.Actor.Pid())
		}
		return true
	})
	msg := packet_msg.Pack2Msg(&pt.HeartReq{})
	for _, name := range pidList {
		server.RpcActor.Rpc2Actor(name, msg)
	}
}
