package server

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"net"
	pt2 "simple_game/game/api/protos/pt"
	"simple_game/game/pkg"
)

type server struct {
	pt2.UnimplementedHelloServer
}

func (s *server) Say(ctx context.Context, req *pt2.SayRequest) (*pt2.SayResponse, error) {
	fmt.Println("request:", req.Msg)
	return &pt2.SayResponse{Msg: "Hello " + req.Msg}, nil
}

func StartGRPC() {
	listen, err := net.Listen("tcp", ":9901")
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}

	s := grpc.NewServer()
	pt2.RegisterHelloServer(s, &server{})
	defer func() {
		s.Stop()
		_ = listen.Close()
	}()
	pkg.INFO("grpc server-start-", 9901)
	err = s.Serve(listen)
	if err != nil {
		pkg.ERROR("failed to serve", err)
		return
	}

}
