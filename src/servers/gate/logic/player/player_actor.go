package player

import (
	"common/rpcutil"
	"framework/actor"
	"framework/log"
	"framework/proto/pb/ssplayer"
	"framework/rpc"
)

type NetAttrPlayerId struct{}

type PlayerActor struct {
	playerId int64
	connId   int64
}

// 玩家登入
type ReqLogin struct {
}

type RespLogin struct {
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
	case *ReqLogin:
		player.login(ctx, req)
	case *LogoutReq:
		player.logout(ctx, req)
	}
}

// 玩家登入
func (player *PlayerActor) login(ctx actor.Context, req *ReqLogin) {
	log.Info("player login, playerId=%v", player.playerId)
	reqLogin := &ssplayer.ReqPlayerLogin{}
	reqLogin.PlayerId = new(int64)
	*reqLogin.PlayerId = player.playerId
	respLogin := &ssplayer.RespPlayerLogin{}
	rpcutil.CallPlayer(player.playerId, reqLogin, respLogin)

	ctx.Respond(&RespLogin{})
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

	reqLogout := &ssplayer.ReqPlayerLogout{}
	reqLogout.PlayerId = new(int64)
	*reqLogout.PlayerId = player.playerId
	respLogout := &ssplayer.RespPlayerLogout{}
	rpcutil.CallPlayer(player.playerId, reqLogout, respLogout)

	ctx.Respond(&LogoutResp{})
}
