package playermgr

type Player struct {
	id           int64
	netSessionId int64
}

func NewPlayer(playerId int64) *Player {
	return &Player{
		id: playerId,
	}
}

func (player *Player) GetId() int64 {
	return player.id
}

func (player *Player) SetNetSessionId(sessionId int64) {
	player.netSessionId = sessionId
}

func (player *Player) GetNetSessionId() int64 {
	return player.netSessionId
}
