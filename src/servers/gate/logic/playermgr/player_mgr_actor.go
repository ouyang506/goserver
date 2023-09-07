package playermgr

import (
	"framework/actor"
	"framework/log"
	"framework/network"
	"gate/logic/player"
)

var (
	ActorName = "player_mgr_actor"
	ActorId   = actor.NewActorId(ActorName)
)

type PlayerMgrActor struct {
	playerMap map[int64]*actor.ActorID
	connMap   map[int64]*actor.ActorID
}

// 添加玩家
type AddPlayerReq struct {
	PlayerId int64
	NetConn  network.Connection
}

type AddPlayerResp struct {
	PlayerActorId *actor.ActorID
}

// 查询玩家
type FindPlayerReq struct {
	PlayerId int64
}

type FindPlayerResp struct {
	PlayerActorId *actor.ActorID
}

// 删除玩家
type RemovePlayerReq struct {
	PlayerId int64
}

type RemovePlayerResp struct {
	PlayerActorId *actor.ActorID
}

func NewPlayerMgrActor() *PlayerMgrActor {
	return &PlayerMgrActor{}
}

// 管理所有玩家，不要阻塞
func (playerMgr *PlayerMgrActor) Receive(ctx actor.Context) {
	switch req := ctx.Message().(type) {
	case *actor.Start:
		playerMgr.playerMap = make(map[int64]*actor.ActorID)
		playerMgr.connMap = make(map[int64]*actor.ActorID)
	case *AddPlayerReq:
		playerMgr.addPlayer(ctx, req)
	case *FindPlayerReq:
		playerMgr.findPlayer(ctx, req)
	case *RemovePlayerReq:
		playerMgr.removePlayer(ctx, req)
	}
}

// 添加玩家
func (mgr *PlayerMgrActor) addPlayer(ctx actor.Context, req *AddPlayerReq) {
	playerId := req.PlayerId
	playerActorId, ok := mgr.playerMap[playerId]
	if !ok {
		playerActorId = ctx.Spawn(player.NewPlayerActor(playerId))
		mgr.playerMap[playerId] = playerActorId
		log.Info("player manager add player, playerId=%v, actorId=%v", playerId, playerActorId)
	}

	req.NetConn.SetAttrib(player.NetAttrPlayerId{}, playerId)
	ctx.Respond(&AddPlayerResp{PlayerActorId: playerActorId})
}

// 查找玩家
func (mgr *PlayerMgrActor) findPlayer(ctx actor.Context, req *FindPlayerReq) {
	playerId := req.PlayerId
	playerActorId, ok := mgr.playerMap[playerId]
	if !ok {
		ctx.Respond(&FindPlayerResp{PlayerActorId: nil})
		return
	}
	ctx.Respond(&FindPlayerResp{PlayerActorId: playerActorId})
}

// 删除玩家
// notice : 删除时将停止掉PlayerActor,删除后不要再往PlayerActor发送消息
func (mgr *PlayerMgrActor) removePlayer(ctx actor.Context, req *RemovePlayerReq) {
	playerId := req.PlayerId
	playerActorId, ok := mgr.playerMap[playerId]
	if !ok {
		ctx.Respond(&RemovePlayerResp{PlayerActorId: nil})
		return
	}

	log.Info("player manager remove player, playerId=%v, actorId=%v", playerId, playerActorId)

	delete(mgr.playerMap, playerId)
	ctx.Stop(playerActorId)

	ctx.Respond(&RemovePlayerResp{PlayerActorId: playerActorId})
}
