package fsm

type Event struct {
	Name     string
	Src      string
	Dst      string
	ExtParam []any
}

type EventTransition struct {
	Name string
	Src  []string
	Dst  string
}

type EventTransKey struct {
	event string
	src   string
}

type Callback func(*Event)

type CallbackType int

const (
	CbTypeNone        CallbackType = 0
	CbTypeLeaveState  CallbackType = 1
	CbTypeEnterState  CallbackType = 2
	CbTypeTickState   CallbackType = 3
	CbTypeChangeState CallbackType = 4
)

type CallbackKey struct {
	cbType  CallbackType
	cbState string
}
