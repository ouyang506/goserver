package robot

import (
	"fmt"
	"framework/log"
	"mockclient/config"
	"mockclient/netmgr"
	"time"

	"utility/fsm"
	"utility/queue"
)

// fsm state
const (
	StateLogout       = "state_logout"
	StateLoginAccount = "state_login_account"
	StateLoginGate    = "state_login_gate"
	StateQueryPlayer  = "state_query_player"
)

// fsm event
const (
	EventLogout         = "event_logout"
	EventLoginAccount   = "event_login_account"
	EventLoginGate      = "event_login_gate"
	EventQueryPlayer    = "event_query_player"
	EventGateDisconnect = "event_get_disconnect"
)

const (
	RobotUserName = "admin"
	RobotPassword = "123456"
)

type Robot struct {
	fsm             *fsm.FSM
	asyncEventQueue *queue.LockFreeQueue

	netmgr *netmgr.NetMgr

	conf     *config.Config
	gateIp   string
	gatePort int
	token    string

	lastFailTime int64 // millseconds
}

func NewRobot(conf *config.Config) *Robot {
	robot := &Robot{
		conf:            conf,
		asyncEventQueue: queue.NewLockFreeQueue(),
	}
	robot.fsm = fsm.NewFSM(StateLogout,
		[]fsm.EventTransition{
			{Name: EventLoginAccount, Src: []string{StateLogout}, Dst: StateLoginAccount},
		},
		map[string]fsm.Callback{
			"tick_" + StateLogout:       robot.tickStateLogout,
			"tick_" + StateLoginAccount: robot.tickStateLoginAccount,
		})
	return robot
}

// run in one goroutine
func (robot *Robot) Update() {
	robot.fsm.Update()
	for {
		v := robot.asyncEventQueue.Dequeue()
		if v == nil {
			break
		}
		event := v.([]any)[0].(string)
		extParam := v.([]any)[0].([]any)
		robot.fsm.Event(event, extParam...)
	}
}

func (robot *Robot) AsyncPostEvent(event string, extParam ...any) {
	robot.asyncEventQueue.Enqueue([]any{event, extParam})
}

func (robot *Robot) checkLastFailTime(delta int64) bool {
	if robot.lastFailTime <= 0 {
		return true
	}
	return time.Now().UnixMilli() > robot.lastFailTime+delta
}

func (robot *Robot) resetLastFailTime() {
	robot.lastFailTime = 0
}

func (robot *Robot) tickStateLogout(e *fsm.Event) {
	if !robot.checkLastFailTime(200) {
		return
	}

	if len(robot.conf.LoginServers.LoginServer) <= 0 {
		log.Error("login addr conf error")
		robot.lastFailTime = time.Now().UnixMilli()
		return
	}

	ip := robot.conf.LoginServers.LoginServer[0].IP
	port := robot.conf.LoginServers.LoginServer[0].Port
	url := fmt.Sprintf("http://%s:%d/login", ip, port)
	err, loginResp := httpLogin(url, RobotUserName, RobotPassword)
	if err != nil {
		log.Error("http login error: %s", err)
		robot.lastFailTime = time.Now().UnixMilli()
		return
	}

	robot.resetLastFailTime()

	robot.gateIp = loginResp.GateIp
	robot.gatePort = loginResp.GatePort
	robot.token = loginResp.Token

	robot.fsm.Event(EventLoginGate)
}

func (robot *Robot) tickStateLoginAccount(e *fsm.Event) {
	if !robot.checkLastFailTime(200) {
		return
	}

	if robot.netmgr == nil {
		robot.netmgr = netmgr.NewNetMgr()
		robot.netmgr.Start()
	}

	if !robot.netmgr.FindGateStub(robot.gateIp, robot.gatePort) {
		robot.netmgr.RemoveGateStubs()
		robot.netmgr.AddGateStub(robot.gateIp, robot.gatePort)
	}

	robot.resetLastFailTime()
	robot.fsm.Event(EventLoginAccount)
}
