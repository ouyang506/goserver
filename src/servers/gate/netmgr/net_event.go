package netmgr

import (
	"common"
	"framework/log"
	"framework/network"
	"framework/proto/pb"
	"framework/proto/pb/cscommon"
	"framework/rpc"
	"gate/logic/handler"
	"gate/logic/player"
)

// 网络事件回调
type ClientNetEventHandler struct {
	msgHandler *handler.MessageHandler
	rpc.OuterNetEventHandler
}

func NewClientNetEventHandler(msgHandler *handler.MessageHandler) *ClientNetEventHandler {
	return &ClientNetEventHandler{
		msgHandler: msgHandler,
	}
}

func (e *ClientNetEventHandler) OnAccept(c network.Connection) {
	e.OuterNetEventHandler.OnAccept(c)
}

func (e *ClientNetEventHandler) OnConnect(c network.Connection, err error) {
	//服务器不会主动去连接客户端
	log.Error("unexpected network event OnConnect")
}

func (e *ClientNetEventHandler) OnClosed(c network.Connection) {
	e.OuterNetEventHandler.OnClosed(c)
	go e.msgHandler.OnNetConnClosed(c)
}

func (e *ClientNetEventHandler) OnRcvMsg(c network.Connection, msg interface{}) {
	outerMsg := msg.(*rpc.OuterMessage)
	msgId := outerMsg.MsgID
	callId := outerMsg.CallId

	// handled by gate server
	if msg == 0 || (msgId > int(cscommon.CsMsgRoute_cs_gate_msg_id_begin) &&
		msgId <= int(cscommon.CsMsgRoute_cs_gate_msg_id_end)) {
		e.OuterNetEventHandler.OnRcvMsg(c, msg)
		return
	}

	// check login status
	data, ok := c.GetAttrib(player.NetAttrPlayerId{})
	if !ok {
		log.Error("player not login, message is not allowed, msgId=%v", msgId)
		return
	}

	playerId, ok := data.(int64)
	if !ok || playerId <= 0 {
		log.Error("load connection attribute player id error, data=%v, sessionId=%v",
			data, c.GetSessionId())
		return
	}

	targetServer := 0
	switch {
	case msgId >= int(cscommon.CsMsgRoute_cs_player_msg_id_begin) &&
		msgId <= int(cscommon.CsMsgRoute_cs_player_msg_id_end):
		{
			targetServer = common.ServerTypePlayer
		}
	}

	if targetServer == 0 {
		log.Error("route client message error, msgId=%v", msgId)
		return
	}

	//TODO : 这里proto message进行了多次不必要的反序列化和序列化，codec需要重新规划
	go func() {
		reqMsg, respMsg := pb.GetProtoMsgById(msgId)
		if reqMsg == nil {
			log.Error("cannot find the rpc message, msgId=%v", msgId)
			return
		}
		if respMsg == nil {
			err := rpc.Notify(targetServer, playerId, outerMsg.Content)
			if err != nil {
				log.Error("route to target server error: %v", err)
				return
			}
		} else {
			err := rpc.Call(targetServer, playerId, outerMsg.Content, respMsg)
			if err != nil {
				log.Error("route to target server error: %v", err)
				return
			}
			respOuterMsg := &rpc.OuterMessage{}
			respOuterMsg.CallId = callId
			respOuterMsg.MsgID = 0
			respOuterMsg.Content = respMsg
			rpc.TcpSend(rpc.RpcModeOuter, c.GetSessionId(), respOuterMsg)
		}
	}()
}
