package playermgr

import (
	"framework/actor"
	"framework/log"
	"framework/rpc"
	"gate/logic/player"
)

var (
	ActorName = "player_mgr_actor"
	ActorId   = actor.NewActorId(ActorName)
)

// 添加玩家
type ReqAddPlayer struct {
	PlayerId int64
	ConnId   int64
}

type RespAddPlayer struct {
	PlayerActorId *actor.ActorID
}

// 删除玩家
type ReqRemovePlayer struct {
	PlayerId int64
}

type RespRemovePlayer struct {
	PlayerActorId *actor.ActorID
}

// 查询玩家
type ReqFindPlayer struct {
	PlayerId int64
}

type RespFindPlayer struct {
	PlayerActorId *actor.ActorID
}

// 通过网络连接查询玩家
type ReqFindPlayerByConn struct {
	ConnId int64
}

type RespFindPlayerByConn struct {
	PlayerId      int64
	PlayerActorId *actor.ActorID
}

// 通过网络连接删除玩家
type RemovePlayerByConnReq struct {
	ConnId int64
}

type RemovePlayerByConnResp struct {
	PlayerId      int64
	PlayerActorId *actor.ActorID
}

type PlayerInfo struct {
	playerId int64
	actorId  *actor.ActorID
	connId   int64
}

type PlayerMgrActor struct {
	playerMap map[int64]*PlayerInfo
	connMap   map[int64]*PlayerInfo
}

func NewPlayerMgrActor() *PlayerMgrActor {
	return &PlayerMgrActor{
		playerMap: make(map[int64]*PlayerInfo),
		connMap:   make(map[int64]*PlayerInfo),
	}
}

// 管理所有玩家，不要阻塞
func (playerMgr *PlayerMgrActor) Receive(ctx actor.Context) {
	switch req := ctx.Message().(type) {
	case *ReqAddPlayer:
		playerMgr.addPlayer(ctx, req)
	case *ReqRemovePlayer:
		playerMgr.removePlayer(ctx, req)
	case *ReqFindPlayer:
		playerMgr.findPlayer(ctx, req)
	case *ReqFindPlayerByConn:
		playerMgr.findPlayerByConn(ctx, req)
	case *RemovePlayerByConnReq:
		playerMgr.removePlayerByConn(ctx, req)
	}
}

// 添加玩家
func (mgr *PlayerMgrActor) addPlayer(ctx actor.Context, req *ReqAddPlayer) {
	playerId := req.PlayerId
	playerInfo, ok := mgr.playerMap[playerId]
	if !ok {
		playerActorId := ctx.Spawn(player.NewPlayerActor(playerId, req.ConnId))
		playerInfo = &PlayerInfo{
			playerId: playerId,
			actorId:  playerActorId,
			connId:   req.ConnId,
		}
		mgr.playerMap[playerId] = playerInfo
		mgr.connMap[req.ConnId] = playerInfo
		log.Info("player manager add player, playerId=%v, actorId=%v, connId=%v",
			playerId, playerActorId, req.ConnId)
	} else {
		// 关闭旧连接
		oldConnId := playerInfo.connId
		// 赋值新的connId
		playerInfo.connId = req.ConnId

		delete(mgr.connMap, oldConnId)
		mgr.connMap[req.ConnId] = playerInfo

		log.Info("close player old net connection, playerId=%v, oldConnId=%v, newConnId=%v",
			playerId, oldConnId, req.ConnId)
		rpc.TcpClose(rpc.RpcModeOuter, oldConnId)
	}

	ctx.Respond(&RespAddPlayer{PlayerActorId: playerInfo.actorId})
}

// 删除玩家
// notice : 删除时将停止掉PlayerActor,删除后不要再往PlayerActor发送消息
func (mgr *PlayerMgrActor) removePlayer(ctx actor.Context, req *ReqRemovePlayer) {
	playerId := req.PlayerId
	playerInfo, ok := mgr.playerMap[playerId]
	if !ok {
		ctx.Respond(&RespRemovePlayer{PlayerActorId: nil})
		return
	}

	log.Info("player manager remove player, playerId=%v, actorId=%v", playerId, playerInfo.actorId)

	delete(mgr.playerMap, playerId)
	delete(mgr.connMap, playerInfo.connId)
	ctx.Stop(playerInfo.actorId)

	ctx.Respond(&RespRemovePlayer{PlayerActorId: playerInfo.actorId})
}

// 查找玩家
func (mgr *PlayerMgrActor) findPlayer(ctx actor.Context, req *ReqFindPlayer) {
	playerId := req.PlayerId
	playerInfo, ok := mgr.playerMap[playerId]
	if !ok {
		ctx.Respond(&RespFindPlayer{PlayerActorId: nil})
		return
	}
	ctx.Respond(&RespFindPlayer{PlayerActorId: playerInfo.actorId})
}

// 通过网络连接查找玩家
func (mgr *PlayerMgrActor) findPlayerByConn(ctx actor.Context, req *ReqFindPlayerByConn) {
	connId := req.ConnId
	playerInfo, ok := mgr.connMap[connId]
	if !ok {
		ctx.Respond(&RespFindPlayerByConn{PlayerId: 0, PlayerActorId: nil})
		return
	}
	ctx.Respond(&RespFindPlayerByConn{PlayerId: playerInfo.playerId, PlayerActorId: playerInfo.actorId})
}

// 通过网络连接删除玩家
func (mgr *PlayerMgrActor) removePlayerByConn(ctx actor.Context, req *RemovePlayerByConnReq) {
	connId := req.ConnId
	playerInfo, ok := mgr.connMap[connId]
	if !ok {
		ctx.Respond(&RemovePlayerByConnResp{PlayerActorId: nil})
		return
	}

	log.Info("player manager remove player by connection, connId=%v, playerId=%v, playerActorId=%v",
		connId, playerInfo.playerId, playerInfo.actorId)

	delete(mgr.playerMap, playerInfo.playerId)
	delete(mgr.connMap, connId)
	ctx.Stop(playerInfo.actorId)

	ctx.Respond(&RemovePlayerByConnResp{PlayerId: playerInfo.playerId, PlayerActorId: playerInfo.actorId})
}
