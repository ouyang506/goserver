package actor

import (
	"utility/safemap"

	murmur32 "github.com/twmb/murmur3"
)

// actor系统
type ActorSystem struct {
	root       *ContextImpl
	mailboxMap *safemap.ConcurrentMap[string, *MailBox]
}

func NewActorSystem() *ActorSystem {
	sys := &ActorSystem{}
	sys.root = newContextImpl(sys, nil, rootActorID)
	sys.mailboxMap = safemap.NewConcurrentMap[string, *MailBox](1024,
		func(k string) uint32 {
			return murmur32.Sum32([]byte(k))
		})
	return sys
}

func (system *ActorSystem) Root() Context {
	return system.root
}

func (system *ActorSystem) addMailBox(actorId *ActorID, mb *MailBox) bool {
	_, ok := system.mailboxMap.Get(actorId.Key())
	if ok {
		return false
	}
	system.mailboxMap.Set(actorId.Key(), mb)
	return true
}

func (system *ActorSystem) getMailBox(actorId *ActorID) *MailBox {
	mb, ok := system.mailboxMap.Get(actorId.Key())
	if !ok {
		return nil
	}
	return mb
}

func (system *ActorSystem) removeMailBox(actorId *ActorID) {
	system.mailboxMap.Del(actorId.Key())

}

func (system *ActorSystem) Spawn(parent *ActorID, actor Actor) *ActorID {
	return system.SpawnNamed(parent, "", actor)
}

func (system *ActorSystem) SpawnNamed(parent *ActorID, name string, actor Actor) *ActorID {
	actorId := genActorId(name)
	mb := newMailBox(system, parent, actorId, actor)
	ok := system.addMailBox(actorId, mb)
	if !ok {
		return nil
	}

	// 发送start消息给用户
	sysMsg := &SystemMessage{
		message: systemStart,
	}
	system.SendSysMsg(actorId, sysMsg)

	return actorId
}

func (system *ActorSystem) SendSysMsg(target *ActorID, msg *SystemMessage) {
	mb := system.getMailBox(target)
	if mb == nil {
		return
	}
	mb.pushSysMsg(msg)
}

func (system *ActorSystem) SendUserMsg(sender *ActorID, target *ActorID, msg *UserMessage) {
	mb := system.getMailBox(target)
	if mb == nil {
		return
	}
	mb.pushUserMsg(msg)
}

func (system *ActorSystem) Stop(actorId *ActorID) {
	mb := system.getMailBox(actorId)
	if mb == nil {
		return
	}
	mb.stop()
	system.removeMailBox(actorId)
}
