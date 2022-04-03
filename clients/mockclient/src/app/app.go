package app

import (
	"common/consts"
	"common/log"
	"common/pbmsg"
	"common/rpc"
	"fmt"
	"mockclient/config"
	"mockclient/netmgr"
	"os"
	"sync"
	"time"
)

var (
	appOnce      = sync.Once{}
	app     *App = nil
)

func GetApp() *App {
	appOnce.Do(func() {
		app = new(App)
		app.init()
	})
	return app
}

type App struct {
	conf   *config.Config
	netMgr *netmgr.NetMgr
}

func (app *App) init() bool {
	// init config
	app.conf = config.NewConfig()
	err := app.conf.Load("conf/config.xml")
	if err != nil {
		fmt.Printf("Load config error, program exit! error message : %v", err)
		os.Exit(1)
	}

	//init logger
	logger := log.NewCommonLogger()
	logger.AddSink(log.NewStdLogSink())
	logger.AddSink(log.NewFileLogSink("../log/", log.RotateByHour))
	logger.Start()
	log.SetLogger(logger)

	// init net manager
	app.netMgr = netmgr.NewNetMgr()
	app.netMgr.Init(app.conf)

	return true
}

func (app *App) Start() {
	log.Info("====================================================================")
	log.Info("")
	log.Info("App start ..")

	// log app config info
	log.Info("App config info :")
	log.Info("%+v", app.conf)

	app.netMgr.Start()

	for {
		app.update()
		time.Sleep(time.Second * 10000)
	}
}

func (app *App) update() {

	req := &pbmsg.LoginGateReqT{}
	req.Username = new(string)
	*req.Username = "test_name"
	req.Passwd = new(string)
	*req.Passwd = "test_passwd"

	log.Debug("begin rpc call")
	resp, err := rpc.Call(consts.ServerTypeGate, int(pbmsg.MsgID_login_gate_req), req)
	log.Debug("end rpc call, error: %v, resp: %v", err, resp)
}
