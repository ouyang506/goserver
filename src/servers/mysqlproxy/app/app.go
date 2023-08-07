package app

import (
	"fmt"
	"framework/log"
	"mysqlproxy/config"
	"mysqlproxy/netmgr"
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
	err := app.conf.Load("../../../conf/mysqlproxy.xml")
	if err != nil {
		fmt.Printf("Load config error, program exit! error message : %v", err)
		os.Exit(1)
	}

	//init logger
	logger := log.NewCommonLogger()
	logger.AddSink(log.NewStdLogSink())
	logger.AddSink(log.NewFileLogSink("mysqlproxy", "../../../logs/", log.RotateByHour))
	logger.Start()
	log.SetLogger(logger)

	// init net manager
	app.netMgr = netmgr.NewNetMgr(app.conf)
	app.netMgr.Init(app.conf)

	return true
}

func (app *App) Start() {
	log.Info("********************************************************")
	log.Info("**************     mysqlproxy server   *****************")
	log.Info("********************************************************")
	log.Info("App Start ..")
	// log app config info
	log.Info("App config info :")
	log.Info("%+v", app.conf)

	app.netMgr.Start()

	for {
		app.update()
		time.Sleep(time.Millisecond * 20)
	}
}

// main loop
func (app *App) update() {

}
