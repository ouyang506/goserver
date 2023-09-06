package actor

import (
	"errors"
	"sync/atomic"
	"utility/queue"
)

const (
	ScheduleIdle int32 = iota
	ScheduleRunning
)

const (
	normal int32 = iota
	stopping
	stopped
)

type MailBox struct {
	system  *ActorSystem
	actorId *ActorID
	actor   Actor
	ctx     *ContextImpl

	sysMsgQueue  *queue.LockFreeQueue
	userMsgQueue *queue.LockFreeQueue

	scheduleStatus int32
	stopStatus     int32
}

func newMailBox(system *ActorSystem, parent *ActorID, aid *ActorID, actor Actor) *MailBox {
	mb := &MailBox{
		system:         system,
		actorId:        aid,
		actor:          actor,
		sysMsgQueue:    queue.NewLockFreeQueue(),
		userMsgQueue:   queue.NewLockFreeQueue(),
		scheduleStatus: ScheduleIdle,
		stopStatus:     normal,
	}
	mb.ctx = newContextImpl(system, parent, aid)
	return mb
}

func (mb *MailBox) stop() {
	if atomic.CompareAndSwapInt32(&mb.stopStatus, normal, stopping) {
		sysMsg := &SystemMessage{
			message: systemStop,
		}
		mb.pushSysMsg(sysMsg)
	}
}

func (mb *MailBox) active() bool {
	return atomic.LoadInt32(&mb.stopStatus) == normal
}

func (mb *MailBox) sysMsgCount() int32 {
	return mb.sysMsgQueue.Length()
}

func (mb *MailBox) pushSysMsg(msg *SystemMessage) {
	mb.sysMsgQueue.Enqueue(msg)
	mb.schedule()
}

func (mb *MailBox) popSysMsg() *SystemMessage {
	v := mb.sysMsgQueue.Dequeue()
	if v == nil {
		return nil
	}
	return v.(*SystemMessage)
}

func (mb *MailBox) userMsgCount() int32 {
	return mb.userMsgQueue.Length()
}

func (mb *MailBox) pushUserMsg(msg *UserMessage) {
	mb.userMsgQueue.Enqueue(msg)
	mb.schedule()
}

func (mb *MailBox) popUserMsg() *UserMessage {
	v := mb.userMsgQueue.Dequeue()
	if v == nil {
		return nil
	}
	return v.(*UserMessage)
}

func (mb *MailBox) schedule() {
	if atomic.CompareAndSwapInt32(&mb.scheduleStatus, ScheduleIdle, ScheduleRunning) {
		go mb.process()
	}
}

func (mb *MailBox) process() {
run:
	for {
		// 将所有系统消息全部处理完再处理用户消息
		sysMsg := mb.popSysMsg()
		if sysMsg != nil {
			mb.ctx.msg = sysMsg
			mb.actor.Receive(mb.ctx)
			mb.ctx.msg = nil

			if _, ok := sysMsg.Message().(*Stop); ok {
				for {
					userMsg := mb.popUserMsg()
					if userMsg == nil {
						break
					}
					if userMsg.Future() != nil {
						userMsg.Future().cancel(errors.New("actor stopped"))
					}
				}
				return
			}

			continue
		}

		userMsg := mb.popUserMsg()
		if userMsg != nil {
			mb.ctx.msg = userMsg
			mb.actor.Receive(mb.ctx)
			mb.ctx.msg = nil
			continue
		}

		break
	}

	atomic.StoreInt32(&mb.scheduleStatus, ScheduleIdle)

	// 设置idle前新进来msg，状态为running，process不会调用到
	// 重新判断queue长度以免遗漏消息，cas状态以保证只有一个协程在处理
	if mb.sysMsgQueue.Length() > 0 || mb.userMsgQueue.Length() > 0 {
		if atomic.CompareAndSwapInt32(&mb.scheduleStatus, ScheduleIdle, ScheduleRunning) {
			// go mb.process()
			goto run
		}
	}
}
