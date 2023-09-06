package actor

type Start struct{}
type Stop struct{}

var (
	// actor创建后发送的系统消息
	systemStart = &Start{}
	// actor调用stop后发送的系统消息,丢弃余的消息
	systemStop = &Stop{}
)

type Message interface {
	Message() any
	Sender() *ActorID
	Future() *Future
}

// 系统消息
type SystemMessage struct {
	message any
}

func (msg *SystemMessage) Message() any {
	return msg.message
}

func (msg *SystemMessage) Sender() *ActorID {
	return nil
}

func (msg *SystemMessage) Future() *Future {
	return nil
}

// 用户消息
type UserMessage struct {
	message any
	sender  *ActorID
	future  *Future
}

func (msg *UserMessage) Message() any {
	return msg.message
}

func (msg *UserMessage) Sender() *ActorID {
	return msg.sender
}

func (msg *UserMessage) Future() *Future {
	return msg.future
}
