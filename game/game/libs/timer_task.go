package libs

import (
	"container/list"
	"simple_game/game/pkg"
	"time"
)

var TickerEventQueue list.List

type TickerEvent struct {
	Id         int64
	TriggerAt  int64
	ScheduleAt int64
	Callback   func(interface{})
}

func AddTimer(triggerAt int64, call func(interface{})) {
	now := time.Now().Unix()
	if now >= triggerAt {
		// 直接触发
		call(nil)
	}
	event := &TickerEvent{
		Id:        pkg.NewUid(),
		TriggerAt: triggerAt,
		Callback:  call,
	}
	TickerEventQueue.PushBack(event)
}

func AddScheduleTimer(scheduleTime int64, call func(interface{})) {
	triggerAt := time.Now().Unix() + scheduleTime
	event := &TickerEvent{
		Id:         pkg.NewUid(),
		TriggerAt:  triggerAt,
		Callback:   call,
		ScheduleAt: scheduleTime,
	}
	TickerEventQueue.PushBack(event)
}

func CheckEventTrigger() {
	if TickerEventQueue.Len() == 0 {
		return
	}
	prev := TickerEventQueue.Front().Prev()
	event := prev.Value.(TickerEvent)
	if event.TriggerAt <= time.Now().Unix() {
		event.Callback(nil)
	}
	TickerEventQueue.Remove(prev)
	//CheckEventTrigger()
}

func TimerStart() {
	for {
		if TickerEventQueue.Len() == 0 {
			continue
		}
		prev := TickerEventQueue.Front()
		if prev == nil {
			continue
		}
		event := prev.Value.(*TickerEvent)
		if event.TriggerAt <= time.Now().Unix() {
			event.Callback(nil)
			TickerEventQueue.Remove(prev)

			// 循环定时器
			if event.ScheduleAt > 0 {
				AddScheduleTimer(event.ScheduleAt, event.Callback)
			}
		}

	}
}

func init() {
	go TimerStart()
}
