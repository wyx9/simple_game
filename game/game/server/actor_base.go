package server

import (
	"context"
	"errors"
	"simple_game/game/pkg"
	"sync/atomic"
	"time"
)

type IActor interface {
	Start()
	Stop()
	Handler(msg interface{}) interface{}
}

var (
	ErrActorNotFound = errors.New("actor not found")
	ErrActorClose    = errors.New("actor close")
	ErrSendTimeout   = errors.New("send timeout")
)

type ActorConfig = struct {
	QueueSize      int
	SendTimeout    time.Duration
	HandlerTimeout time.Duration
}

var DefaultConfig = ActorConfig{
	QueueSize:      1024,
	SendTimeout:    5 * time.Second,
	HandlerTimeout: 5 * time.Second,
}

type ActorBase struct {
	name    string
	iActor  IActor
	request chan interface{}
	stop    chan struct{}
	config  ActorConfig
	closed  int32
	ctx     context.Context
	cancel  context.CancelFunc
}

func (ac *ActorBase) isClosed() bool {
	return atomic.LoadInt32(&ac.closed) == 1
}

func (ac *ActorBase) setClose() {
	atomic.StoreInt32(&ac.closed, 1)
}

func (ac *ActorBase) loop() {

	ac.iActor.Start()

	go func() {
		defer func() {
			ac.iActor.Stop()
			ac.setClose()
			close(ac.request)
			close(ac.stop)

		}()

		for {
			select {
			case <-ac.ctx.Done():
				pkg.INFO("Actor context canceled , name:", ac.name)
				return

			case req, ok := <-ac.request:
				if !ok {
					return
				}

				handleCtx, cancel := context.WithTimeout(ac.ctx, ac.config.HandlerTimeout)
				done := make(chan struct{})

				go func() {
					defer close(done)
					switch r := req.(type) {
					case CallRequest:
						result := ac.iActor.Handler(r.Msg)
						r.Response <- result
					default:
						ac.iActor.Handler(req)
					}
				}()

				select {
				case <-done:
				case <-handleCtx.Done():
					pkg.ERROR("message handling timed out for actor ,", ac.name)
				}
				cancel()
			}
		}
	}()
}

func (ac *ActorBase) Shutdown(timeout time.Duration) error {
	if ac.isClosed() {
		return ErrActorClose
	}
	ac.cancel()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-timer.C:
		return ErrSendTimeout
	case <-ac.stop:
		return nil
	}
}

func NewActorBase(name string, callBack IActor, config *ActorConfig) *ActorBase {
	if config == nil {
		config = &ActorConfig{}
	}

	ctx, cancel := context.WithCancel(context.Background())
	ab := &ActorBase{
		name:    name,
		iActor:  callBack,
		request: make(chan interface{}, config.QueueSize),
		stop:    make(chan struct{}),
		ctx:     ctx,
		cancel:  cancel,
	}

	if err := ActorManner.Add(ab); err != nil {
		return nil
	}

	ab.loop()
	return ab
}
