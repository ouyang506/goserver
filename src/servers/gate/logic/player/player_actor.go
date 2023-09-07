package player

import (
	"framework/actor"
	"framework/log"
	"framework/network"
	"framework/rpc"
)

type NetAttrPlayer struct{}

type PlayerActor struct {
	playerId int64
	netConn  network.Connection
}

type BindNetConnReq struct {
	NetConn network.Connection
}
type BindNetConnResp struct {
}

func NewPlayerActor(playerId int64) *PlayerActor {
	return &PlayerActor{
		playerId: playerId,
	}
}

func (player *PlayerActor) Receive(ctx actor.Context) {
	switch req := ctx.Message().(type) {
	case *actor.Start:
	case *actor.Stop:
	case *BindNetConnReq:
		player.bindNetConn(ctx, req)
	}
}

// 绑定网络连接
func (player *PlayerActor) bindNetConn(ctx actor.Context, req *BindNetConnReq) {
	newNetConn := req.NetConn
	oldNetConn := player.netConn
	oldSessionId := int64(0)
	if oldNetConn != nil {
		oldSessionId = oldNetConn.GetSessionId()
	}

	player.netConn = newNetConn
	player.netConn.SetAttrib(NetAttrPlayer{}, player.playerId)

	log.Info("player bind network connection, playerId=%v, oldSession=%v, newSession=%v",
		player.playerId, oldSessionId, newNetConn.GetSessionId())

	// 关闭旧的连接
	if oldNetConn != nil && oldNetConn != newNetConn {
		oldNetConn.DelAttrib(NetAttrPlayer{})
		rpc.TcpClose(rpc.RpcModeOuter, oldNetConn.GetSessionId())
		log.Info("player bind new connection, so close the old connection, playerId=%v, sessionId=%v",
			player.playerId, oldNetConn.GetSessionId())
	}

	// TODO: broadcast其他gate下线相同的player

	ctx.Respond(&BindNetConnResp{})
}
