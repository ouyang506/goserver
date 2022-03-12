package app

import (
	"common/consts"
	"common/log"
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
	rpc := rpc.CreateRpc()
	rpc.TargetSvrType = int(consts.ServerTypeGate)
	rpc.Request = []byte("this is a rpc request")

	log.Debug("begin rpc call")
	rpc.Call()
	log.Debug("end rpc call, resp: %v", string(rpc.Response))
}
