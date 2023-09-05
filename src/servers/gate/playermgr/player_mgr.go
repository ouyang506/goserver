package playermgr

import (
	"framework/log"
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

func (mgr *PlayerMgr) LoadAndStorePlayer(playerId int64, player *Player) *Player {
	mgr.playerMu.Lock()
	defer mgr.playerMu.Unlock()

	oldPlayer, ok := mgr.playerMap[playerId]
	mgr.playerMap[player.GetId()] = player
	log.Info("add player to map, playerId=%v", playerId)
	if ok {
		return oldPlayer
	} else {
		return nil
	}
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

func (mgr *PlayerMgr) OnDisconnect(playerId int64) {
	mgr.playerMu.RLock()
	defer mgr.playerMu.RUnlock()

	player, ok := mgr.playerMap[playerId]
	if !ok {
		log.Error("network disconnected, but player not found, playerId=%v", playerId)
		return
	}

	delete(mgr.playerMap, playerId)
	mgr.OnLogout(player)
}

func (mgr *PlayerMgr) OnLogin(player *Player) {
	log.Info("player login, playerId=%v", player.GetId())
}

func (mgr *PlayerMgr) OnLogout(player *Player) {
	log.Info("player logout, playerId=%v", player.GetId())
}
