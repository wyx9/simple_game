package controller

import (
	"simple_game/server"
)

type BaseController struct {
}

func (this BaseController) Ctx(ctx interface{}) *server.Player {
	return ctx.(*server.Player)
}
