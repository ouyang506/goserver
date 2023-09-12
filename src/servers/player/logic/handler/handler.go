package handler

import (
	"framework/actor"
)

// handler会被任意协程调用，勿必保证handler线程安全
type MessageHandler struct {
	rootContext actor.Context
}

func NewMessageHandler(rootContext actor.Context) *MessageHandler {
	return &MessageHandler{
		rootContext: rootContext,
	}
}

func (h *MessageHandler) Root() actor.Context {
	return h.rootContext
}
