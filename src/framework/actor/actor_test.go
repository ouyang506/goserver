package actor

import (
	"fmt"
	"testing"
	"time"
)

type MyActor struct {
}

type HelloReq struct {
	content string
}

type HelloResp struct {
	content string
}

func (*MyActor) Receive(ctx Context) {
	switch req := ctx.Message().(type) {
	case *Start:
		fmt.Println("started")
	case *Stop:
		fmt.Println("stopped")
	case *HelloReq:
		fmt.Println(req.content)
		ctx.Respond(&HelloResp{"hello response"})
	}
}

func TestActor(t *testing.T) {
	root := NewActorSystem().Root()
	actorId := root.Spawn(&MyActor{})
	future := root.Request(actorId, &HelloReq{"hello request"})
	result, err := future.WaitTimeout(time.Second * 3)
	if err != nil {
		t.Fatalf("err = %v\n", err)
	} else {
		fmt.Printf("result = %v\n", result.(*HelloResp).content)
	}

	root.Stop(actorId)
	time.Sleep(2 * time.Second)
}
