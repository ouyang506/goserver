package handler

import (
	"framework/log"
	"framework/network"
	"gate/logic/player"
	"gate/logic/playermgr"
	"time"
)

// 客户端网络断开
func (h *MessageHandler) OnNetConnClosed(conn network.Connection) {
	// 查询Player
	findReq := &playermgr.FindPlayerByConnReq{
		ConnId: conn.GetSessionId(),
	}
	future := h.Root().Request(playermgr.ActorId, findReq)
	findResp, err := future.WaitTimeout(time.Second * 5)
	if err != nil {
		log.Error("client disconnected, find player error, connId=%v, err=%v", conn.GetSessionId(), err)
	}
	playerId := findResp.(*playermgr.FindPlayerByConnResp).PlayerId
	playerActorId := findResp.(*playermgr.FindPlayerByConnResp).PlayerActorId

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

	// 通过connection删除player
	removeReq := &playermgr.RemovePlayerByConnReq{
		ConnId: conn.GetSessionId(),
	}
	future = h.Root().Request(playermgr.ActorId, removeReq)
	_, err = future.WaitTimeout(time.Second * 5)
	if err != nil {
		log.Error("client disconnected, remove player error, playerId=%v, connId=%v, err=%v",
			playerId, conn.GetSessionId(), err)
	}
}
