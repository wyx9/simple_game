package controller

import (
	"google.golang.org/protobuf/proto"
	"simple_game/game/api/protos/pt"
	"simple_game/game/server"
	"time"
)

type AllController struct {
	BaseController
}

func LoginController(actor *server.Player, req *pt.LoginReq) *pt.LoginRes {
	return &pt.LoginRes{
		UUid: req.UUid,
		Code: 0,
	}
}

func TestController(actor *server.Player, req *pt.HeartReq) proto.Message {
	return &pt.HeartRes{Time: time.Now().Unix()}
}
