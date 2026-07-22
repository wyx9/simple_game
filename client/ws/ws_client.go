package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"nhooyr.io/websocket"

	"simple_game/game/api/protos/pt"
	"simple_game/game/libs"
)

// WebSocket 测试客户端，对应服务端 ws listener（默认 ws://127.0.0.1:8889/ws）。
// 消息格式与 TCP 客户端一致：JSON(PacketMsg{Name, Data})，承载在 WS 文本帧中。
//
// 运行： go run ./client/ws

const (
	serverURL        = "ws://127.0.0.1:8889/ws"
	heartbeatPeriod  = 3 * time.Second  // 心跳间隔
	clientTimeout    = 30 * time.Second // 心跳超时
)

var (
	lastHeartbeatResp int64
	isConnected       int32
)

func main() {
	for {
		if err := connectAndServe(); err != nil {
			fmt.Println("连接失败:", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if atomic.LoadInt32(&isConnected) == 0 {
			continue
		}
	}
}

func connectAndServe() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, serverURL, nil)
	if err != nil {
		return fmt.Errorf("连接服务器失败: %v", err)
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	fmt.Println("已连接到服务器")
	atomic.StoreInt32(&isConnected, 1)
	atomic.StoreInt64(&lastHeartbeatResp, time.Now().Unix())

	// 读取协程
	go reader(c)
	// 心跳协程
	go sendHeartbeat(c)

	// 主循环：从控制台读取输入，发送 LoginReq
	for atomic.LoadInt32(&isConnected) == 1 {
		var msg string
		_, _ = fmt.Scanln(&msg)

		if atomic.LoadInt32(&isConnected) == 0 {
			break
		}

		req := &pt.LoginReq{
			UUid:     12312312,
			Name:     msg,
			PassWord: "123",
		}
		packBy := libs.Pack2Msg(req)

		// WS 文本帧承载 JSON(PacketMsg)，与服务端 wsConn.WriteMessage 一致
		writeCtx, writeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := c.Write(writeCtx, websocket.MessageText, packBy); err != nil {
			fmt.Println("发送消息失败:", err)
			atomic.StoreInt32(&isConnected, 0)
			writeCancel()
			break
		}
		writeCancel()
	}
	return nil
}

func reader(c *websocket.Conn) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
		_, data, err := c.Read(ctx)
		cancel()
		if err != nil {
			atomic.StoreInt32(&isConnected, 0)
			fmt.Println("连接断开:", err)
			return
		}
		atomic.StoreInt64(&lastHeartbeatResp, time.Now().Unix())
		fmt.Println("收到消息：" + string(data))
	}
}

func sendHeartbeat(c *websocket.Conn) {
	ticker := time.NewTicker(heartbeatPeriod)
	defer ticker.Stop()

	for range ticker.C {
		if atomic.LoadInt32(&isConnected) == 0 {
			return
		}

		heartbeat := &pt.HeartReq{}
		packBy := libs.Pack2Msg(heartbeat)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := c.Write(ctx, websocket.MessageText, packBy); err != nil {
			fmt.Println("发送心跳包失败:", err)
			cancel()
			continue
		}
		cancel()

		if time.Now().Unix()-atomic.LoadInt64(&lastHeartbeatResp) > int64(clientTimeout.Seconds()) {
			fmt.Println("心跳超时，断开连接")
			atomic.StoreInt32(&isConnected, 0)
			_ = c.Close(websocket.StatusNormalClosure, "")
			return
		}
	}
}
