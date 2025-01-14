package server

import (
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"net"
	"simple_game/game/libs"
	"simple_game/game/pkg"
	"simple_game/game/routes"
	"time"
)

const tickDuration = 10 * time.Second

type Player struct {
	name       string
	ip         string
	ac         IActor
	conn       *net.Conn
	PlayerData *PlayerData
	timer      *time.Timer
}

type PlayerData struct {
	Name        string
	LastLoginAt int64
}

func (p *Player) Start() {
	p.loadPlayer()
	p.scheduleNextTick()
}

func (p *Player) Stop() {
	// 持久化
	if p.timer != nil {
		p.timer.Stop()
	}
	pkg.INFO("player stop ,player name :", p.name)
	p.PersistenceKey()
}

func (p *Player) PersistenceKey() string {
	return fmt.Sprintf("player_data_%s", p.PlayerData.Name)
}
func (p *Player) persistence() {
	data := p.PlayerData
	marshal, _ := json.Marshal(data)
	key := p.PersistenceKey()
	pkg.RedisClient.Set(pkg.RCtx, key, marshal, -1)
	pkg.INFO("player persistence", p.name)
}
func (p *Player) loadPlayer() {
	key := p.PersistenceKey()
	cmd := pkg.RedisClient.Get(pkg.RCtx, key)
	if cmd.Val() == "" {
		return
	}
	playerData := p.PlayerData

	_ = json.Unmarshal([]byte(cmd.Val()), playerData)
	pkg.INFO("player persistence", p.name)
}
func (p *Player) Handler(msg interface{}) interface{} {
	switch msg {
	case "tick":
		p.onTick()
	default:
		p.HandlerByClient(msg)
	}
	return nil
}

func (p *Player) scheduleNextTick() {
	time.AfterFunc(tickDuration, func() {
		err := ActorManner.CastMsg(p.name, "tick")
		if err != nil {
			return
		}
		p.scheduleNextTick()
	})
}

func (p *Player) onTick() {
	p.persistence()
}

func (p *Player) HandlerByClient(msg interface{}) {
	bytes, ok := msg.([]byte)
	if !ok {
		return
	}
	codePack := libs.DeCodePack(bytes)
	if codePack == nil {
		return
	}
	handler, res, err := routes.Route(p, codePack.Name, codePack.Data)
	if err != nil {
		pkg.ERROR("handler is nil , ", handler)
	} else {
		pkg.DEBUG("req :", codePack.Name)
		marshal, _ := protojson.Marshal(res)
		pkg.DEBUG("res :", proto.MessageName(res).Name(), string(marshal))
		c := *p.conn
		_, _ = c.Write(marshal)
	}
}

func StartNewPlayerActor(name string, conn net.Conn) (*Player, error) {
	playerActor := &Player{
		name: name,
		ip:   conn.RemoteAddr().String(),
		conn: &conn,
		PlayerData: &PlayerData{
			Name:        name,
			LastLoginAt: time.Now().Unix(),
		},
	}
	_ = NewActorBase(name, playerActor, nil)
	return playerActor, nil
}
