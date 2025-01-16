package register

import (
	"google.golang.org/protobuf/proto"
	"simple_game/api/protos/pt"
	"simple_game/controller"
	"simple_game/routes"
)

func RegisteredRoute() {
	allController := controller.AllController{}

	routes.AddRoute("LoginReq", func(ctx interface{}, req any) proto.Message {
		player := allController.Ctx(ctx)
		return controller.LoginController(player, req.(*pt.LoginReq))
	})

	// 测试
	routes.AddRoute("HeartReq", func(ctx interface{}, req any) proto.Message {
		player := allController.Ctx(ctx)
		return controller.TestController(player, req.(*pt.HeartReq))
	})

}
