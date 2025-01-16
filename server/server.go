package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"simple_game/api/protos/pt"
	"simple_game/config"
	"simple_game/libs"
	"simple_game/pkg"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

// 服务器相关常量定义
const (
	maxConnections   = 1000             // 最大连接数限制
	heartbeatTimeout = 30 * time.Second // 心跳超时时间
	shutdownTimeout  = 10 * time.Second // 服务器关闭超时时间
	readTimeout      = 10 * time.Second // 读取超时时间
)

// Server 游戏服务器结构体
type Server struct {
	Ip       string                 // 服务器IP地址
	port     int                    // 服务器监听端口
	connMap  map[string]*Connection // 连接管理映射表，key为连接ID
	connLock sync.RWMutex           // 连接映射表的读写锁
	listener net.Listener           // TCP监听器
	ctx      context.Context        // 服务器上下文
	cancel   context.CancelFunc     // 上下文取消函数
	wg       sync.WaitGroup         // 等待组，用于优雅关闭
}

// Connection 客户端连接结构体
type Connection struct {
	uuid       string             // 连接唯一标识符
	conn       net.Conn           // TCP连接对象
	lastActive time.Time          // 最后活跃时间
	ctx        context.Context    // 连接上下文
	cancel     context.CancelFunc // 连接取消函数
	isAuth     bool               // 是否已认证
	actorId    string             // 关联的Actor ID
	session    string             // 会话标识
}

// NewServer 创建新的服务器实例
func NewServer(ip string, port int) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	server := &Server{
		Ip:      ip,
		port:    port,
		connMap: make(map[string]*Connection),
		ctx:     ctx,
		cancel:  cancel,
	}
	return server
}

// GlobalServer 全局服务器实例
var GlobalServer *Server

// Start 启动游戏服务器
func Start() {
	// 从配置中获取端口
	port, err := strconv.Atoi(config.Conf.Tcp.Port)
	if err != nil {
		return
	}
	server := NewServer(config.Conf.Tcp.Addr, port)
	GlobalServer = server

	// 创建TCP监听器
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, port))
	if err != nil {
		return
	}

	server.listener = listener

	server.wg.Add(1)
	go server.serverLoop()
}

// serverLoop 服务器主循环，处理新的连接请求
func (s *Server) serverLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			// 接受新的客户端连接
			conn, err := s.listener.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of close network connection") {
					pkg.ERROR("accept server error", err)
				}
				continue
			}
			// TODO: 检查是否超过最大连接数

			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

// handleConnection 处理单个客户端连接
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	connId := conn.RemoteAddr().String()
	ctx, cancel := context.WithCancel(s.ctx)

	// 创建新的连接实例
	connection := &Connection{
		uuid:       conn.RemoteAddr().String(),
		conn:       conn,
		lastActive: time.Now(),
		ctx:        ctx,
		cancel:     cancel,
	}

	// 添加连接到连接映射表
	if !s.addConnection(connId, connection) {
		_ = conn.Close()
		cancel()
		return
	}

	defer func() {
		// TODO: 清理session
		s.removeConnection(connId)
		_ = conn.Close()
		cancel()
	}()

	// 主消息循环
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 设置读取超时
			err := conn.SetReadDeadline(time.Now().Add(readTimeout))
			if err != nil {
				return
			}

			// 接收客户端数据
			buf, err := pkg.RecvData(conn)
			if err != nil {
				return
			}

			if err != nil {
				ActorManner.FindAndClosePlayer(connId)
				return
			}

			// 更新最后活跃时间
			connection.lastActive = time.Now()

			// 处理接收到的消息
			err = s.handleMessage(connection, buf, conn)
			if err != nil {
				pkg.ERROR("handleMessage error : ", err)
			}
		}
	}
}

// handleMessage 处理接收到的消息
func (s *Server) handleMessage(c *Connection, buf []byte, conn net.Conn) error {
	msg := string(buf)
	pkg.INFO("handleMessage ,", msg)

	// 解码消息包
	codePack := libs.DeCodePack(buf)
	if codePack == nil {
		return errors.New("decode pack error")
	}

	// 未认证的连接只允许处理登录请求
	if !c.isAuth {
		if codePack.Name != "LoginReq" {
			return errors.New("no is auth")
		}

		// 处理登录请求
		temp := &pt.LoginReq{}
		err := proto.Unmarshal(codePack.Data, temp)
		if err != nil {
			return errors.New("decode error")
		}

		// TODO: 创建session

		c.isAuth = true
		c.actorId = strconv.Itoa(int(temp.GetUUid()))

		// 启动新的玩家Actor
		_, err = StartNewPlayerActor(c.actorId, conn)
		if err != nil {
			return errors.New("start player actor error")
		}
	}
	// TODO: 认证成功的actor进行校验
	// TODO: 验证用户的Id是否匹配

	return ActorManner.CastMsg(c.actorId, buf)
}

// removeConnection 从连接映射表中移除连接
func (s *Server) removeConnection(connId string) {
	s.connLock.Lock()
	defer s.connLock.Unlock()
	delete(s.connMap, connId)
}

// addConnection 添加连接到连接映射表
func (s *Server) addConnection(connId string, connection *Connection) bool {
	s.connLock.Lock()
	defer s.connLock.Unlock()
	if _, exists := s.connMap[connId]; exists {
		return false
	}
	s.connMap[connId] = connection
	return true
}

// StopServer 停止服务器
func StopServer() error {
	if GlobalServer == nil {
		return nil
	}

	// 触发关闭信号
	GlobalServer.cancel()

	// 关闭监听器
	if GlobalServer.listener != nil {
		err := GlobalServer.listener.Close()
		if err != nil {
			return err
		}
	}

	// 等待所有goroutine完成
	done := make(chan struct{})
	go func() {
		GlobalServer.wg.Wait()
		close(done)
	}()

	// 等待关闭完成或超时
	select {
	case <-done:
		pkg.INFO("server shutdown completed gracefully ")
	case <-time.After(shutdownTimeout):
		pkg.ERROR("server shutdown timed out")
	}
	return nil
}
