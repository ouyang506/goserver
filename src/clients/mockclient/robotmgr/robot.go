package robotmgr

import (
	"common/rpcutil"
	"fmt"
	"framework/log"
	"framework/proto/pb/cs"
	"mockclient/configmgr"
	"mockclient/netmgr"
	"time"

	"math/rand"
	"utility/fsm"
	"utility/queue"
)

// fsm state
const (
	StateLogout        = "state_logout"
	StateLoginAccount  = "state_login_account"
	StateCreateAccount = "state_create_account"
	StateCreateRole    = "state_create_role"
	StateLoginGate     = "state_login_gate"
	StateQueryPlayer   = "state_query_player"
)

// fsm event
const (
	EventLogout         = "event_logout"
	EventLoginAccount   = "event_login_account"
	EventCreateAccount  = "event_create_account"
	EventCreateRole     = "event_create_role"
	EventLoginGate      = "event_login_gate"
	EventQueryPlayer    = "event_query_player"
	EventGateDisconnect = "event_get_disconnect"
)

const (
	RobotUsername = "mock_robot"
	RobotPassword = "123456"
	RobotNickname = "mock_robot_nick"
)

type Robot struct {
	fsm             *fsm.FSM
	asyncEventQueue *queue.LockFreeQueue

	stateFrameTime int64 // millseconds

	gateIp   string
	gatePort int
	token    string
	roleId   int64
}

func newRobot() *Robot {
	robot := &Robot{}
	robot.fsm = fsm.NewFSM(StateLogout,
		[]fsm.EventTransition{
			{Name: EventLoginAccount, Src: []string{StateLogout}, Dst: StateLoginAccount},
			{Name: EventCreateAccount, Src: []string{StateLoginAccount}, Dst: StateCreateAccount},
			{Name: EventCreateRole, Src: []string{StateLoginAccount, StateCreateAccount}, Dst: StateCreateRole},
			{Name: EventLoginGate, Src: []string{StateLoginAccount, StateCreateRole}, Dst: StateLoginGate},
			{Name: EventQueryPlayer, Src: []string{StateLoginGate}, Dst: StateQueryPlayer},
		},
		map[string]fsm.Callback{
			"change_state":               robot.ChangeState,
			"tick_" + StateLogout:        robot.tickStateLogout,
			"tick_" + StateLoginAccount:  robot.tickStateLoginAccount,
			"tick_" + StateCreateAccount: robot.tickStateCreateAccount,
			"tick_" + StateCreateRole:    robot.tickStateCreateRole,
			"tick_" + StateLoginGate:     robot.tickStateLoginGate,
			"tick_" + StateQueryPlayer:   robot.tickStateQueryPlayer,
		})
	robot.Reset()
	return robot
}

func (robot *Robot) Reset() {
	robot.asyncEventQueue = queue.NewLockFreeQueue()
	robot.stateFrameTime = 0
	robot.gateIp = ""
	robot.gatePort = 0
	robot.token = ""
	robot.roleId = 0
}

// run in one goroutine
func (robot *Robot) Update() {
	for {
		v := robot.asyncEventQueue.Dequeue()
		if v == nil {
			break
		}
		event := v.([]any)[0].(string)
		extParam := v.([]any)[0].([]any)
		robot.fsm.Event(event, extParam...)
	}

	robot.fsm.Update()
}
func (robot *Robot) ChangeState(e *fsm.Event) {
	log.Info("robot fsm change state, event = %v, source = %v, dest = %v",
		e.Name, e.Src, e.Dst)
}

func (robot *Robot) Event(event string) {
	robot.resetStateFrameTime()
	if robot.fsm.CanEvent(event) {
		robot.fsm.Event(event)
	} else {
		log.Error("can not post fsm event, curent state: %v, event : %v", robot.fsm.CurrentState(), event)
	}
}

func (robot *Robot) AsyncPostEvent(event string, extParam ...any) {
	robot.asyncEventQueue.Enqueue([]any{event, extParam})
}

func (robot *Robot) checkStateFrameTime(delta int64) bool {
	if robot.stateFrameTime <= 0 {
		return true
	}
	now := time.Now().UnixMilli()
	if now < robot.stateFrameTime+delta {
		return false
	}
	robot.stateFrameTime = now
	return true
}

func (robot *Robot) resetStateFrameTime() {
	robot.stateFrameTime = 0
}

func (robot *Robot) tickStateLogout(e *fsm.Event) {
	robot.Reset()
	robot.Event(EventLoginAccount)
}

// 登录账号
func (robot *Robot) tickStateLoginAccount(e *fsm.Event) {
	if !robot.checkStateFrameTime(500) {
		return
	}

	conf := configmgr.Instance().GetConfig()
	if len(conf.LoginServers.LoginServer) <= 0 {
		log.Error("login addr conf error")
		robot.resetStateFrameTime()
		return
	}

	idx := rand.Int() % len(conf.LoginServers.LoginServer)
	ip := conf.LoginServers.LoginServer[idx].IP
	port := conf.LoginServers.LoginServer[idx].Port
	url := fmt.Sprintf("http://%s:%d/account/login", ip, port)
	loginAccountResp, err := loginAccount(url, RobotUsername, RobotPassword)
	if err != nil {
		log.Error("login account error: %s", err)
		robot.resetStateFrameTime()
		return
	}

	if loginAccountResp.ErrCode != 0 {
		//	ErrCodeInvalidUserOrPasswd = 103
		//  ErrDescInvalidUserOrPasswd = "login user name or password error"
		//  没有该账号则创建账号
		if loginAccountResp.ErrCode == 103 {
			robot.Event(EventCreateAccount)
			return
		} else {
			log.Error("login account resp error code: %v, desc: %v", loginAccountResp.ErrCode,
				loginAccountResp.ErrDesc)
			robot.resetStateFrameTime()
			return
		}
	}

	// 没有角色则创建角色
	if loginAccountResp.RoleId == 0 {
		robot.Event(EventCreateRole)
		return
	}

	robot.gateIp = loginAccountResp.GateIp
	robot.gatePort = loginAccountResp.GatePort
	robot.token = loginAccountResp.Token
	robot.roleId = loginAccountResp.RoleId
	robot.Event(EventLoginGate)
}

// 创建账号
func (robot *Robot) tickStateCreateAccount(e *fsm.Event) {
	if !robot.checkStateFrameTime(500) {
		return
	}

	conf := configmgr.Instance().GetConfig()
	if len(conf.LoginServers.LoginServer) <= 0 {
		log.Error("login addr conf error")
		robot.resetStateFrameTime()
		return
	}

	idx := rand.Int() % len(conf.LoginServers.LoginServer)
	ip := conf.LoginServers.LoginServer[idx].IP
	port := conf.LoginServers.LoginServer[idx].Port
	url := fmt.Sprintf("http://%s:%d/account/create", ip, port)
	createAccountResp, err := createAccount(url, RobotUsername, RobotPassword)
	if err != nil {
		log.Error("create account error: %s", err)
		robot.resetStateFrameTime()
		return
	}

	if createAccountResp.ErrCode != 0 {
		log.Error("create account resp error code: %v, desc: %v", createAccountResp.ErrCode,
			createAccountResp.ErrDesc)
		robot.resetStateFrameTime()
		return
	}
	robot.Event(EventCreateRole)
}

// 创建角色
func (robot *Robot) tickStateCreateRole(e *fsm.Event) {
	if !robot.checkStateFrameTime(500) {
		return
	}

	conf := configmgr.Instance().GetConfig()
	if len(conf.LoginServers.LoginServer) <= 0 {
		log.Error("login addr conf error")
		robot.resetStateFrameTime()
		return
	}

	idx := rand.Int() % len(conf.LoginServers.LoginServer)
	ip := conf.LoginServers.LoginServer[idx].IP
	port := conf.LoginServers.LoginServer[idx].Port
	url := fmt.Sprintf("http://%s:%d/role/create", ip, port)
	createRoleResp, err := createRole(url, RobotUsername, RobotPassword, RobotNickname)
	if err != nil {
		log.Error("create role error: %s", err)
		robot.resetStateFrameTime()
		return
	}

	if createRoleResp.ErrCode != 0 {
		log.Error("create role resp error code: %v, desc: %v", createRoleResp.ErrCode,
			createRoleResp.ErrDesc)
		robot.resetStateFrameTime()
		return
	}

	robot.gateIp = createRoleResp.GateIp
	robot.gatePort = createRoleResp.GatePort
	robot.token = createRoleResp.Token
	robot.roleId = createRoleResp.RoleId
	robot.Event(EventLoginGate)
}

// 登录gate
func (robot *Robot) tickStateLoginGate(e *fsm.Event) {
	if !robot.checkStateFrameTime(500) {
		return
	}

	netmgr.Instance().CheckStart()

	netmgr.Instance().RemoveGateStubs()
	netmgr.Instance().AddGateStub(robot.gateIp, robot.gatePort)

	req := &cs.ReqLoginGate{}
	resp := &cs.RespLoginGate{}
	req.RoleId = new(int64)
	*req.RoleId = robot.roleId
	req.Token = new(string)
	*req.Token = robot.token
	err := rpcutil.ClientCall(req, resp)
	if err != nil {
		log.Error("rpc call login gate failed: %v", err)
		robot.resetStateFrameTime()
		return
	}

	if resp.GetErrCode() != 0 {
		log.Error("rpc call login gate response error code: %v, desc: %v",
			resp.GetErrCode(), resp.GetErrDesc())
		robot.resetStateFrameTime()
		return
	}

	robot.Event(EventQueryPlayer)
}

// 查询玩家
func (robot *Robot) tickStateQueryPlayer(e *fsm.Event) {
	if !robot.checkStateFrameTime(500) {
		return
	}
}
