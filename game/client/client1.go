package main

import (
	"fmt"
	"net"
	"os"
	"simple_game/game/api/protos/pt"
	"simple_game/game/libs"
	"simple_game/game/pkg"
	"time"
)

var HeartTime int64

const ClientTimeOut = time.Minute * 10

// 客户端读取消息
func Reader(conn *net.TCPConn) {
	buff := make([]byte, 1280)
	for {
		j, err := conn.Read(buff)
		if err != nil {
			//ch <- 1
			break
		}
		fmt.Println("收到消息：" + string(buff[:j]))
		HeartTime = time.Now().Unix()
	}
}

func StartClient(network, address string) {
	tcpAddr, _ := net.ResolveTCPAddr(network, address)
	conn, err := net.DialTCP(network, nil, tcpAddr)
	if err != nil {
		fmt.Println("tcp conn fail ->", err)
		os.Exit(0)
	}
	fmt.Println(" 已经连接上服务器")
	go Reader(conn)
	defer func() {
		_ = conn.Close()
	}()
	HeartTime = time.Now().Unix()
	for {
		// 发送消息
		var msg string
		_, _ = fmt.Scanln(&msg)
		// 发送的消息

		// todo 测试 转换成pt
		req := pt.LoginReq{
			UUid:     12312312,
			Name:     msg,
			PassWord: "123",
		}
		packBy := libs.Pack2Msg(&req)
		// 封装协议并且发送
		if err := pkg.SendData(conn, packBy); err != nil {
			fmt.Println("客户端发送消息,封装失败")
		}

		if time.Now().Unix()-HeartTime > int64(ClientTimeOut) {
			fmt.Println("连接超时断开连接")
			os.Exit(0)
		}
	}
}

func main() {
	StartClient("tcp", "127.0.0.1:8888")

	//RpcClient()
}
