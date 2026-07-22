package controller

import (
	"simple_game/game"
)

type BaseController struct {
}

func (this BaseController) Ctx(ctx interface{}) *game.Player {
	return ctx.(*game.Player)
}
