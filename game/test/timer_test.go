package test

import (
	"context"
	"fmt"
	"testing"
	"time"
)

//var Mutex = sync.Once{}

func process() {
	duration := 5 * time.Second
	timer := time.NewTicker(duration)
	//ctx, _ := context.WithTimeout(ctx, 2*time.Second)
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	//Mutex.Do(func() {
	//
	//})
	go processLogic(timer, ctx, func() {
		fmt.Println("定时器触发")
		//Mutex.Unlock()
	})

	for {
		return
	}
}

func processLogic(timer *time.Ticker, ctx context.Context, callBack func()) {
	defer timer.Stop()
	for {
		//
		//timer.Reset(duration)
		select {
		case <-ctx.Done():
			fmt.Println("超时")
			return
		case <-timer.C:
			callBack()
			return
		}
	}
}

func Test_Timer(t *testing.T) {
	process()
}

func Test_Time_After(t *testing.T) {
	time.AfterFunc(time.Second*4, func() {
		fmt.Println(11111111111111)
	})
}
