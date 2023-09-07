package actor

// actor context 接口定义
// 在actor的recieve方法中传递过去
type Context interface {
	Self() *ActorID   // actor的唯一id
	Parent() *ActorID // 产生此actor的父actor
	Message() any     // 用户具体消息
	Sender() *ActorID // 发送者actor

	Spawn(actor Actor) *ActorID                   // 创建新的actor
	SpawnNamed(name string, actor Actor) *ActorID // 指定actor名称来创建

	Send(target *ActorID, message any) error      // 给actor发送用户消息
	Request(target *ActorID, message any) *Future // 发送请求等待返回
	Respond(message any)                          // 返回消息给发送方

	Stop(*ActorID) // 停止并移除actor，不会立即清除
}

type ContextImpl struct {
	system *ActorSystem
	id     *ActorID
	parent *ActorID
	msg    Message
}

func newContextImpl(system *ActorSystem, parent *ActorID, id *ActorID) *ContextImpl {
	ctx := &ContextImpl{
		system: system,
		parent: parent,
		id:     id,
	}
	return ctx
}

func (ctx *ContextImpl) Self() *ActorID {
	return ctx.id
}

func (ctx *ContextImpl) Parent() *ActorID {
	return ctx.parent
}

func (ctx *ContextImpl) Message() any {
	return ctx.msg.Message()
}

func (ctx *ContextImpl) Sender() *ActorID {
	return ctx.msg.Sender()
}

func (ctx *ContextImpl) Spawn(actor Actor) *ActorID {
	return ctx.SpawnNamed("", actor)
}

func (ctx *ContextImpl) SpawnNamed(name string, actor Actor) *ActorID {
	childId := ctx.system.SpawnNamed(ctx.Self(), name, actor)
	return childId
}

func (ctx *ContextImpl) Send(target *ActorID, message any) error {
	sender := ctx.Self()
	userMsg := &UserMessage{
		sender:  sender,
		future:  nil,
		message: message,
	}
	return ctx.system.SendUserMsg(sender, target, userMsg)
}

func (ctx *ContextImpl) Request(target *ActorID, message any) *Future {
	sender := ctx.Self()
	userMsg := &UserMessage{
		sender:  sender,
		future:  newFuture(),
		message: message,
	}
	err := ctx.system.SendUserMsg(sender, target, userMsg)
	if err != nil {
		userMsg.future.cancel(err)
	}
	return userMsg.future
}

func (ctx *ContextImpl) Respond(message any) {
	if ctx.msg == nil {
		return
	}
	future := ctx.msg.Future()
	if future == nil {
		return
	}
	future.respond(message)
}

func (ctx *ContextImpl) Stop(actorId *ActorID) {
	ctx.system.Stop(actorId)
}
