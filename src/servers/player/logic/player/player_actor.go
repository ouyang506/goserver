package player

import (
	"common/mysqlutil"
	"framework/actor"
	"framework/log"
	"time"
)

type PlayerActor struct {
	playerId       int64
	loginTime      time.Time
	lastLoginTime  time.Time
	lastLogoutTime time.Time
}

func NewPlayerActor(playerId int64) *PlayerActor {
	return &PlayerActor{
		playerId: playerId,
	}
}

func (player *PlayerActor) Receive(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Start:
		player.login()
	case *actor.Stop:
		player.logout()
	}
}

func (player *PlayerActor) login() {
	player.loginTime = time.Now()
	player.refreshLoginTime()
	player.load()
}

func (player *PlayerActor) logout() {
	player.refreshLogoutTime()
}

func (player *PlayerActor) load() {
	log.Debug("load player info, playerId=%v", player.playerId)

	row, err := mysqlutil.QueryOne("select * from t_player where id=?", player.playerId)
	if err != nil {
		log.Error("query player error: %v", err)
		return
	}

	if row == nil {
		log.Error("query player nil, playerId=%v", player.playerId)
		return
	}

	lastLoginTime, err := time.ParseInLocation("2006-01-02 15:04:05", row.FieldString("login_time"), time.Local)
	if err != nil {
		player.lastLoginTime = lastLoginTime
	}

	lastLogoutTime, err := time.ParseInLocation("2006-01-02 15:04:05", row.FieldString("logout_time"), time.Local)
	if err != nil {
		player.lastLogoutTime = lastLogoutTime
	}
}

func (player *PlayerActor) refreshLoginTime() {
	_, _, err := mysqlutil.Execute("update t_player set login_time=? where id=?",
		player.lastLoginTime.Format("2006-01-02 15:04:05"), player.playerId)
	if err != nil {
		log.Error("refresh player login time error: %v", err)
	}
}

func (player *PlayerActor) refreshLogoutTime() {
	_, _, err := mysqlutil.Execute("update t_player set logout_time=? where id=?",
		time.Now().Format("2006-01-02 15:04:05"), player.playerId)
	if err != nil {
		log.Error("refresh player login time error: %v", err)
	}
}
