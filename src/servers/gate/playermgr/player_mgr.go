package playermgr

import (
	"framework/log"
	"framework/rpc"
	"sync"
)

var (
	once                 = sync.Once{}
	playerMgr *PlayerMgr = nil
)

// singleton
func Instance() *PlayerMgr {
	once.Do(func() {
		playerMgr = newPlayerMgr()
	})
	return playerMgr
}

// thread safe
type PlayerMgr struct {
	playerMap map[int64]*Player
	playerMu  sync.RWMutex
}

func newPlayerMgr() *PlayerMgr {
	mgr := &PlayerMgr{}
	return mgr
}

func (mgr *PlayerMgr) AddPlayer(player *Player) {
	if player == nil {
		log.Error("player is nil")
		return
	}
	log.Info("add player to map, playerId=%v, netSessionId=%v",
		player.GetId(), player.GetNetSessionId())

	mgr.playerMu.Lock()
	defer mgr.playerMu.Unlock()

	oldPlayer, ok := mgr.playerMap[player.GetId()]
	if ok {
		log.Info("add player to map but player has been existed, playerId=%v, netSessionId=%v",
			player.GetId(), player.GetNetSessionId())
		rpc.TcpClose(rpc.RpcModeOuter, oldPlayer.GetNetSessionId())
	}
	mgr.playerMap[player.GetId()] = player
}

func (mgr *PlayerMgr) RemovePlayer(playerId int64) {
	log.Info("delete player from map, playerId=%v", playerId)
	mgr.playerMu.Lock()
	defer mgr.playerMu.Unlock()

	delete(mgr.playerMap, playerId)
}

func (mgr *PlayerMgr) GetPlayer(playerId int64) *Player {
	mgr.playerMu.RLock()
	defer mgr.playerMu.RUnlock()

	player, ok := mgr.playerMap[playerId]
	if !ok {
		return nil
	}
	return player
}
