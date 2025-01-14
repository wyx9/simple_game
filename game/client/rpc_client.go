package main

import (
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	pt2 "simple_game/game/api/protos/pt"
)

func RpcClient() {
	var serviceHost = "127.0.0.1:9901"

	conn, err := grpc.Dial(serviceHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	client := pt2.NewHelloClient(conn)
	rsp, err := client.Say(context.TODO(), &pt2.SayRequest{
		Msg: "RPC-TEST",
	})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(rsp)

	fmt.Println("按回车键退出程序...")
	in := bufio.NewReader(os.Stdin)
	_, _, _ = in.ReadLine()

}
