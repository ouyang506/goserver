package playermgr

import (
	"framework/actor"
	"framework/network"
	"gate/logic/player"
)

var (
	ActorName = "player_mgr_actor"
	ActorId   = actor.NewActorId(ActorName)
)

type PlayerMgrActor struct {
	playerMap map[int64]*actor.ActorID
}

type PlayerLoginReq struct {
	PlayerId int64
	NetConn  network.Connection
}

type PlayerLoginResp struct {
}

func NewPlayerMgrActor() *PlayerMgrActor {
	return &PlayerMgrActor{}
}

func (playerMgr *PlayerMgrActor) Receive(ctx actor.Context) {
	switch req := ctx.Message().(type) {
	case *actor.Start:
		playerMgr.playerMap = make(map[int64]*actor.ActorID)
	case *PlayerLoginReq:
		playerMgr.playerLogin(ctx, req)
	}
}

// 玩家登陆
func (mgr *PlayerMgrActor) playerLogin(ctx actor.Context, req *PlayerLoginReq) {
	playerId := req.PlayerId
	playerActorId, ok := mgr.playerMap[playerId]
	if !ok {
		playerActorId = ctx.Spawn(player.NewPlayerActor(playerId))
	}

	bindReq := &player.BindNetConnReq{
		NetConn: req.NetConn,
	}
	ctx.Request(playerActorId, bindReq).Wait()

	mgr.playerMap[playerId] = playerActorId
	ctx.Respond(&PlayerLoginResp{})
}
