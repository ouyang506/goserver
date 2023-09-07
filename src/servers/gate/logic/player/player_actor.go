package player

import (
	"framework/actor"
	"framework/log"
	"framework/network"
	"framework/rpc"
)

type NetAttrPlayerId struct{}

type PlayerActor struct {
	playerId int64
	netConn  network.Connection
}

// 玩家登入
type LoginReq struct {
	NetConn network.Connection
}

type LoginResp struct {
}

// 玩家登出原因
const (
	ReasonDisconnected = "disconnected"
)

// 玩家登出
type LogoutReq struct {
	Reason string
}

type LogoutResp struct {
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
	case *LoginReq:
		player.login(ctx, req)
	case *LogoutReq:
		player.logout(ctx, req)
	}
}

// 玩家登入
func (player *PlayerActor) login(ctx actor.Context, req *LoginReq) {
	player.setNetConn(req.NetConn)

	// TODO: broadcast其他gate下线相同的player
	ctx.Respond(&LoginResp{})
}

// 玩家登出
func (player *PlayerActor) logout(ctx actor.Context, req *LogoutReq) {
	sessionId := int64(0)
	if player.netConn != nil {
		sessionId = player.netConn.GetSessionId()
	}
	log.Info("player logout, playerId=%v, reason=%v, sessionId=%v",
		player.playerId, req.Reason, sessionId)

	if req.Reason != ReasonDisconnected {
		if sessionId != 0 {
			rpc.TcpClose(rpc.RpcModeOuter, sessionId)
		}
	}
	ctx.Respond(&LogoutResp{})
}

// 设置网络
func (player *PlayerActor) setNetConn(conn network.Connection) {
	newNetConn := conn
	oldNetConn := player.netConn
	oldSessionId := int64(0)
	if oldNetConn != nil {
		oldSessionId = oldNetConn.GetSessionId()
	}

	player.netConn = newNetConn

	log.Info("player set connection, playerId=%v, oldSession=%v, newSession=%v",
		player.playerId, oldSessionId, newNetConn.GetSessionId())

	// 关闭旧的连接
	if oldNetConn != nil && oldNetConn != newNetConn {
		oldNetConn.DelAttrib(NetAttrPlayerId{})
		rpc.TcpClose(rpc.RpcModeOuter, oldNetConn.GetSessionId())
		log.Info("player connection changed, so close the old connection, playerId=%v, oldSessionId=%v",
			player.playerId, oldNetConn.GetSessionId())
	}
}
