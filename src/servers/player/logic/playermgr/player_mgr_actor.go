package playermgr

import (
	"framework/actor"
	"framework/log"
	"player/logic/player"
)

var (
	ActorName = "player_mgr_actor"
	ActorId   = actor.NewActorId(ActorName)
)

// 登入
type PlayerLoginReq struct {
	PlayerId int64
}
type PlayerLoginResp struct {
	PlayerActorId *actor.ActorID
}

// 登出
type PlayerLogoutReq struct {
	PlayerId int64
}
type PlayerLogoutResp struct {
}

// 获取Player
type FetchPlayerReq struct {
	PlayerId int64
}

type FetchPlayerResp struct {
	PlayerActorId *actor.ActorID
}

type PlayerMgrActor struct {
	players map[int64]*actor.ActorID
}

func NewPlayerMgrActor() *PlayerMgrActor {
	return &PlayerMgrActor{
		players: make(map[int64]*actor.ActorID),
	}
}

func (playerMgr *PlayerMgrActor) Receive(ctx actor.Context) {
	switch req := ctx.Message().(type) {
	case *PlayerLoginReq:
		playerMgr.playerLogin(ctx, req)
	case *PlayerLogoutReq:
		playerMgr.playerLogout(ctx, req)
	case *FetchPlayerReq:
		playerMgr.fetchPlayer(ctx, req)
	}
}

func (playerMgr *PlayerMgrActor) loadOrCreatePlayer(ctx actor.Context, playerId int64) *actor.ActorID {
	actorId, ok := playerMgr.players[playerId]
	if !ok {
		actorId = ctx.Spawn(player.NewPlayerActor(playerId))
		playerMgr.players[playerId] = actorId
	}
	return actorId
}

func (playerMgr *PlayerMgrActor) playerLogin(ctx actor.Context, req *PlayerLoginReq) {
	log.Info("player login , playerId=%v", req.PlayerId)
	actorId := playerMgr.loadOrCreatePlayer(ctx, req.PlayerId)
	ctx.Respond(&PlayerLoginResp{PlayerActorId: actorId})
}

func (playerMgr *PlayerMgrActor) playerLogout(ctx actor.Context, req *PlayerLogoutReq) {
	log.Info("player logout , playerId=%v", req.PlayerId)
	actorId, ok := playerMgr.players[req.PlayerId]
	if !ok {
		ctx.Respond(&PlayerLogoutResp{})
		return
	}

	delete(playerMgr.players, req.PlayerId)
	ctx.Stop(actorId)

	ctx.Respond(&PlayerLogoutResp{})
}

func (playerMgr *PlayerMgrActor) fetchPlayer(ctx actor.Context, req *FetchPlayerReq) {
	actorId := playerMgr.loadOrCreatePlayer(ctx, req.PlayerId)
	ctx.Respond(&FetchPlayerResp{PlayerActorId: actorId})
}
