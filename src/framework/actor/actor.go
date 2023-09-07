package actor

import (
	"fmt"
	"sync/atomic"
)

// actor接口定义
type Actor interface {
	Receive(ctx Context)
}

var (
	sequence = atomic.Int64{}
)

// actor唯一id
type ActorID struct {
	id string
}

func NewActorId(name string) *ActorID {
	if name == "" {
		return &ActorID{
			id: fmt.Sprintf("$%d", sequence.Add(1)),
		}
	} else {
		return &ActorID{
			id: fmt.Sprintf("#%s", name),
		}
	}
}

func (actorId *ActorID) Key() string {
	return actorId.id
}

// sequence = 0 预留给root
var rootActorID *ActorID = &ActorID{"$0"}

// var rootActor Actor = &EmptyActor{}

// type EmptyActor struct {
// }

// func (*EmptyActor) Receive(ctx Context) {
// }
