package controller

import (
	"google.golang.org/protobuf/proto"
	"simple_game/game/api/protos/pt"
	"simple_game/game"
	"time"
)

type AllController struct {
	BaseController
}

func LoginController(actor *game.Player, req *pt.LoginReq) *pt.LoginRes {
	return &pt.LoginRes{
		UUid: req.UUid,
		Code: 0,
	}
}

func TestController(actor *game.Player, req *pt.HeartReq) proto.Message {
	return &pt.HeartRes{Time: time.Now().Unix()}
}
