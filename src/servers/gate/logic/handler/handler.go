package handler

import (
	"framework/actor"
	"framework/log"
	"framework/network"
	"gate/logic/player"
	"gate/logic/playermgr"
	"time"
)

// handler会被任意协程调用，勿必保证handler线程安全
// 外部协议+内部协议均在此处理
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

// 客户端建立连接
func (h *MessageHandler) OnNetConnAccept(conn network.Connection) {

}

// 客户端网络断开
func (h *MessageHandler) OnNetConnClosed(conn network.Connection) {
	data, ok := conn.GetAttrib(player.NetAttrPlayerId{})
	if !ok {
		log.Info("client connection closed, but player id not bound, sessionId=%v", conn.GetSessionId())
		return
	}

	playerId := data.(int64)
	log.Info("client connection closed, sessionId=%v, playerId=%v", conn.GetSessionId(), playerId)

	var playerActorId *actor.ActorID = nil

	// 查询PlayerActorId
	findReq := &playermgr.FindPlayerReq{
		PlayerId: playerId,
	}
	future := h.Root().Request(playermgr.ActorId, findReq)
	findResp, err := future.WaitTimeout(time.Second * 5)
	if err != nil {
		log.Error("client disconnected, find player error, playerId=%v, err=%v", playerId, err)
	}
	if findResp != nil && findResp.(*playermgr.FindPlayerResp).PlayerActorId != nil {
		playerActorId = findResp.(*playermgr.FindPlayerResp).PlayerActorId
	}

	// player登出
	if playerActorId != nil {
		logoutReq := &player.LogoutReq{
			Reason: player.ReasonDisconnected,
		}
		future = h.Root().Request(playerActorId, logoutReq)
		_, err = future.WaitTimeout(time.Second * 5)
		if err != nil {
			log.Error("client disconnected, request logout error, playerId=%v, err=%v", playerId, err)
		}
	}

	// 删除player
	removeReq := &playermgr.RemovePlayerReq{
		PlayerId: playerId,
	}
	future = h.Root().Request(playermgr.ActorId, removeReq)
	_, err = future.WaitTimeout(time.Second * 5)
	if err != nil {
		log.Error("client disconnected, remove player error, playerId=%v, err=%v", playerId, err)
	}
}
