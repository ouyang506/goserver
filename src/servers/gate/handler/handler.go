package handler

// 外部协议+内部协议均在此处理
type MessageHandler struct {
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}
