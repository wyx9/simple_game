package server

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"net"
	"simple_game/game/api/protos/pt"
	"simple_game/game/config"
	"simple_game/game/libs"
	"simple_game/game/pkg"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxConnections   = 1000
	heartbeatTimeout = 30 * time.Second
	shutdownTimeout  = 5 * time.Second
	readTimeout      = 10 * time.Second
)

type Server struct {
	Ip       string
	port     int
	connMap  map[string]*Connection
	connLock sync.RWMutex
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

type Connection struct {
	uuid       string
	conn       net.Conn
	lastActive time.Time
	ctx        context.Context
	cancel     context.CancelFunc
	isAuth     bool
	actorId    string
	session    string
}

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

var GlobalServer *Server

func Start() {
	port, err := strconv.Atoi(config.Conf.Tcp.Port)
	if err != nil {
		return
	}
	server := NewServer(config.Conf.Tcp.Addr, port)
	GlobalServer = server

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, port))
	if err != nil {
		return
	}

	server.listener = listener

	server.wg.Add(1)
	go server.serverLoop()
}

// serverLoop 处理链接请求
func (s *Server) serverLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of close network connection") {
					pkg.ERROR("accept server error", err)
				}
				continue
			}
			// todo 检查是否超过最大连接数

			s.wg.Add(1)
			go s.handleConnection(conn)

		}
	}
}

// 处理单个客户端的连接
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	connId := conn.RemoteAddr().String()
	ctx, cancel := context.WithCancel(s.ctx)

	connection := &Connection{
		uuid:       conn.RemoteAddr().String(),
		conn:       conn,
		lastActive: time.Now(),
		ctx:        ctx,
		cancel:     cancel,
	}

	if !s.addConnection(connId, connection) {
		_ = conn.Close()
		cancel()
		return
	}

	defer func() {
		// todo 清理session

		// 移除连接
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

			buf, err := pkg.RecvData(conn)
			if err != nil {
				return
			}

			if err != nil {
				ActorManner.FindAndClosePlayer(connId)
				return
			}

			connection.lastActive = time.Now()

			err = s.handleMessage(connection, buf, conn)
			if err != nil {
				pkg.ERROR("handleMessage error : ", err)
			}
		}

	}

}

func (s *Server) handleMessage(c *Connection, buf []byte, conn net.Conn) error {
	msg := string(buf)
	pkg.INFO("handleMessage ,", msg)

	codePack := libs.DeCodePack(buf)
	if codePack == nil {
		return errors.New("decode pack error")
	}

	if !c.isAuth {
		// 只允许登录检查
		if codePack.Name != "LoginReq" {
			return errors.New("no is auth")
		}

		temp := &pt.LoginReq{}
		err := proto.Unmarshal(codePack.Data, temp)
		if err != nil {
			return errors.New("decode error")
		}

		// todo 创建session

		c.isAuth = true
		//c.session = sessionon
		c.actorId = strconv.Itoa(int(temp.GetUUid()))

		_, err = StartNewPlayerActor(c.actorId, conn)
		if err != nil {
			return errors.New("start player actor error")
		}
	}
	// todo 认证成功的actor 进行校验

	// todo 验证用户的Id是否匹配

	return ActorManner.CastMsg(c.actorId, buf)
}

func (s *Server) removeConnection(connId string) {
	s.connLock.Lock()
	defer s.connLock.Unlock()
	delete(s.connMap, connId)
}

func (s *Server) addConnection(connId string, connection *Connection) bool {
	s.connLock.Lock()
	defer s.connLock.Unlock()
	if _, exists := s.connMap[connId]; exists {
		return false
	}
	s.connMap[connId] = connection
	return true
}

func StopServer() error {
	if GlobalServer == nil {
		return nil
	}

	GlobalServer.cancel()

	if GlobalServer.listener != nil {
		err := GlobalServer.listener.Close()
		if err != nil {
			return err
		}
	}

	done := make(chan struct{})

	go func() {
		GlobalServer.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		pkg.INFO("server shutdown completed gracefully ")
	case <-time.After(shutdownTimeout):
		pkg.ERROR("server shutdown timed out")
	}
	return nil
}
