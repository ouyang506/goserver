package fsm

import (
	"strings"
)

// finite state machine
// not thread safe
type FSM struct {
	currentState string
	currentEvent *Event
	transitions  map[EventTransKey]string
	callbacks    map[CallbackKey]Callback
}

func NewFSM(initialSate string, trans []EventTransition,
	callbacks map[string]Callback) *FSM {

	fsm := &FSM{
		currentState: "",
		currentEvent: nil,
		transitions:  make(map[EventTransKey]string),
		callbacks:    make(map[CallbackKey]Callback),
	}
	fsm.currentState = initialSate

	for _, et := range trans {
		for _, src := range et.Src {
			fsm.transitions[EventTransKey{et.Name, src}] = et.Dst
		}
	}

	for k, callback := range callbacks {
		cbState := ""
		cbType := CbTypeNone

		switch {
		case strings.HasPrefix(k, "leave_"): // leave old state
			cbState = strings.TrimPrefix(k, "leave_")
			cbType = CbTypeLeaveState
		case strings.HasPrefix(k, "enter_"): // enter new state
			cbState = strings.TrimPrefix(k, "enter_")
			cbType = CbTypeEnterState
		case strings.HasPrefix(k, "tick_"): // tick by update called
			cbState = strings.TrimPrefix(k, "tick_")
			cbType = CbTypeTickState
		default:
			cbState = ""
			cbType = CbTypeNone
		}

		if cbType != CbTypeNone {
			fsm.callbacks[CallbackKey{cbType, cbState}] = callback
		}
	}

	return fsm
}

func (fsm *FSM) Update() {
	cb, ok := fsm.callbacks[CallbackKey{CbTypeTickState, fsm.currentState}]
	if !ok {
		return
	}
	cb(fsm.currentEvent)
}

func (fsm *FSM) CurrentState() string {
	return fsm.currentState
}

func (fsm *FSM) CanEvent(event string) bool {
	_, ok := fsm.transitions[EventTransKey{event, fsm.currentState}]
	return ok
}

func (fsm *FSM) Event(event string, extParam ...any) {
	dst, ok := fsm.transitions[EventTransKey{event, fsm.currentState}]
	if !ok {
		return
	}

	e := &Event{
		Name:     event,
		Src:      fsm.currentState,
		Dst:      dst,
		ExtParam: extParam,
	}

	cb, ok := fsm.callbacks[CallbackKey{CbTypeLeaveState, e.Src}]
	if ok {
		cb(e)
	}

	cb, ok = fsm.callbacks[CallbackKey{CbTypeEnterState, e.Dst}]
	if ok {
		cb(e)
	}

	fsm.currentEvent = e
	fsm.currentState = dst
}
