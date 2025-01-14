package server

import "net"

type IServer interface {
	Handler(conn net.Conn)
	ListenMessages()
	Start()
}
