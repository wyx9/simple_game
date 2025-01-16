package main

import (
	"fmt"
	"net"
	"simple_game/api/protos/pt"
	"simple_game/libs"
	"simple_game/pkg"
	"sync/atomic"
	"time"
)

// 心跳相关常量
const (
	HeartbeatInterval = 3 * time.Second  // 发送心跳包的间隔
	ClientTimeOut     = 30 * time.Second // 客户端超时时间
)

// 客户端状态
var (
	lastHeartbeatResponse int64 // 最后一次收到心跳响应的时间戳
	isConnected           int32 // 连接状态标志
)

// Reader 客户端读取消息
func Reader(conn *net.TCPConn) {
	defer conn.Close()
	buff := make([]byte, 1280)

	for {
		j, err := conn.Read(buff)
		if err != nil {
			atomic.StoreInt32(&isConnected, 0) // 标记连接已断开
			fmt.Println("连接断开:", err)
			return
		}

		// 更新最后心跳响应时间
		atomic.StoreInt64(&lastHeartbeatResponse, time.Now().Unix())
		fmt.Println("收到消息：" + string(buff[:j]))
	}
}

// sendHeartbeat 发送心跳包
func sendHeartbeat(conn *net.TCPConn) {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for range ticker.C {
		if atomic.LoadInt32(&isConnected) == 0 {
			return
		}

		// 创建心跳请求
		heartbeat := &pt.HeartReq{}
		packBy := libs.Pack2Msg(heartbeat)

		// 发送心跳包
		if err := pkg.SendData(conn, packBy); err != nil {
			fmt.Println("发送心跳包失败:", err)
			continue
		}

		// 检查是否超时
		if time.Now().Unix()-atomic.LoadInt64(&lastHeartbeatResponse) > int64(ClientTimeOut.Seconds()) {
			fmt.Println("心跳超时，断开连接")
			atomic.StoreInt32(&isConnected, 0)
			conn.Close()
			return
		}
	}
}

func StartClient(network, address string) {
	for {
		if err := connectAndServe(network, address); err != nil {
			fmt.Println("连接失败:", err)
			time.Sleep(5 * time.Second) // 重连延迟
			continue
		}

		if atomic.LoadInt32(&isConnected) == 0 {
			continue // 如果连接断开，尝试重连
		}
	}
}

func connectAndServe(network, address string) error {
	tcpAddr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		return fmt.Errorf("解析地址失败: %v", err)
	}

	conn, err := net.DialTCP(network, nil, tcpAddr)
	if err != nil {
		return fmt.Errorf("连接服务器失败: %v", err)
	}

	fmt.Println("已连接到服务器")
	atomic.StoreInt32(&isConnected, 1)
	atomic.StoreInt64(&lastHeartbeatResponse, time.Now().Unix())

	// 启动读取协程
	go Reader(conn)
	// 启动心跳协程
	go sendHeartbeat(conn)

	defer conn.Close()

	// 主消息循环
	for atomic.LoadInt32(&isConnected) == 1 {
		var msg string
		_, _ = fmt.Scanln(&msg)

		if atomic.LoadInt32(&isConnected) == 0 {
			break
		}

		// 创建登录请求
		req := pt.LoginReq{
			UUid:     12312312,
			Name:     msg,
			PassWord: "123",
		}
		packBy := libs.Pack2Msg(&req)

		// 发送消息
		if err := pkg.SendData(conn, packBy); err != nil {
			fmt.Println("发送消息失败:", err)
			atomic.StoreInt32(&isConnected, 0)
			break
		}
	}

	return nil
}

func main() {
	StartClient("tcp", "127.0.0.1:8888")
}
