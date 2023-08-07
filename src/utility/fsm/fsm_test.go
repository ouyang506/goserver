package fsm

import (
	"fmt"
	"testing"
	"time"
)

func TestFsm(t *testing.T) {
	type MyFSM struct {
		fsm *FSM
	}
	myfsm := &MyFSM{}
	myfsm.fsm = NewFSM("closed",
		[]EventTransition{
			{"open", []string{"closed"}, "open"},
			{"close", []string{"open"}, "closed"},
		},
		map[string]Callback{
			"enter_closed": func(e *Event) {
				fmt.Println("enterClosed")
			},
			"leave_closed": func(e *Event) {
				fmt.Println("leaveClosed")
			},
			"tick_closed": func(e *Event) {
				fmt.Println("tickClosed")
				myfsm.fsm.Event("open")
			},
			"enter_open": func(e *Event) {
				fmt.Println("enterOpen")
			},
			"leave_open": func(e *Event) {
				fmt.Println("leaveOpen")
			},
			"tick_open": func(e *Event) {
				fmt.Println("tickOpen")
				myfsm.fsm.Event("close")
			},
		})

	for {
		myfsm.fsm.Update()
		time.Sleep(200 * time.Millisecond)
	}
}
