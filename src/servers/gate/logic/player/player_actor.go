package player

import (
	"framework/actor"
	"framework/log"
	"framework/rpc"
)

type NetAttrPlayerId struct{}

type PlayerActor struct {
	playerId int64
	connId   int64
}

// 玩家登入
type LoginReq struct {
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

func NewPlayerActor(playerId int64, connId int64) *PlayerActor {
	return &PlayerActor{
		playerId: playerId,
		connId:   connId,
	}
}

func (player *PlayerActor) Receive(ctx actor.Context) {
	switch req := ctx.Message().(type) {
	case *actor.Start:
		log.Info("player actor start, playerId=%v, actorId=%v", player.playerId, ctx.Self())
	case *actor.Stop:
		log.Info("player actor stopped, playerId=%v, actorId=%v", player.playerId, ctx.Self())
	case *LoginReq:
		player.login(ctx, req)
	case *LogoutReq:
		player.logout(ctx, req)
	}
}

// 玩家登入
func (player *PlayerActor) login(ctx actor.Context, req *LoginReq) {
	log.Info("player login, playerId=%v", player.playerId)
	ctx.Respond(&LoginResp{})
}

// 玩家登出
func (player *PlayerActor) logout(ctx actor.Context, req *LogoutReq) {
	log.Info("player logout, playerId=%v, reason=%v, connId=%v",
		player.playerId, req.Reason, player.connId)
	if req.Reason != ReasonDisconnected {
		if player.connId != 0 {
			rpc.TcpClose(rpc.RpcModeOuter, player.connId)
		}
	}
	ctx.Respond(&LogoutResp{})
}
