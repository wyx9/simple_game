package server

import (
	"sync"
	"time"
)

type AManager struct {
	actorMap  sync.Map
	playerMap sync.Map
}

var ActorManner = AManager{
	actorMap:  sync.Map{},
	playerMap: sync.Map{},
}

func (am *AManager) Add(i *ActorBase) error {
	am.actorMap.Store(i.name, i)
	player, ok := i.iActor.(*Player)
	if ok {
		am.playerMap.Store(player.ip, i)
	}
	return nil
}

func (am *AManager) Find(name string) *ActorBase {
	value, ok := am.actorMap.Load(name)
	if !ok {
		return nil
	}
	actor, ok := value.(*ActorBase)
	if !ok {
		return nil
	}
	return actor
}

func (am *AManager) FindAndClosePlayer(ip string) {
	value, ok := am.actorMap.Load(ip)
	if !ok {
		return
	}
	actor, ok := value.(*ActorBase)
	if !ok {
		return
	}
	defer am.actorMap.Delete(actor.name)

	actor.iActor.Stop()
}

func (am *AManager) CastMsg(name string, msg interface{}) error {
	actor := am.Find(name)
	if actor == nil {
		return ErrActorNotFound
	}
	if actor.isClosed() {
		return ErrActorClose
	}

	timer := time.NewTimer(actor.config.SendTimeout)
	defer timer.Stop()
	select {
	case actor.request <- msg:
		return nil
	case <-timer.C:
		return ErrSendTimeout
	}
}

type CallRequest struct {
	Msg      interface{}
	Response chan interface{}
}

func (am *AManager) CallMsg(name string, msg interface{}) (interface{}, error) {
	actor := am.Find(name)
	if actor == nil {
		return nil, ErrActorNotFound
	}
	if actor.isClosed() {
		return nil, ErrActorClose
	}

	req := &CallRequest{
		Msg:      msg,
		Response: make(chan interface{}, 1),
	}

	timer := time.NewTimer(actor.config.SendTimeout)
	defer timer.Stop()

	select {
	case actor.request <- req:
	case <-timer.C:
		return nil, ErrSendTimeout
	}

	select {
	case resp := <-req.Response:
		return resp, nil
	case <-timer.C:
		return nil, ErrSendTimeout
	}
}
