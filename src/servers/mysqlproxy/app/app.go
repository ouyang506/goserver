package app

import (
	"fmt"
	"framework/log"
	"mysqlproxy/configmgr"
	"mysqlproxy/mysqlmgr"
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
}

func (app *App) init() bool {
	// init config
	conf := configmgr.Instance().GetConfig()
	if conf == nil {
		fmt.Printf("load config error, program exit!")
		os.Exit(1)
	}

	//init logger
	logger := log.NewCommonLogger()
	logger.SetLogLevel(log.LogLevelDebug)
	logger.AddSink(log.NewStdLogSink())
	logger.AddSink(log.NewFileLogSink("mysqlproxy", "../../../logs/", log.RotateByHour))
	logger.Start()
	log.SetLogger(logger)

	return true
}

func (app *App) Start() {
	log.Info("********************************************************")
	log.Info("**************     mysqlproxy server   *****************")
	log.Info("********************************************************")
	log.Info("App Start ..")
	// log app config info
	log.Info("App config info :")
	log.Info("%+v", configmgr.Instance().GetConfig())

	mysqlmgr.Instance().Start()
	netmgr.Instance().Start()

	for {
		app.update()
		time.Sleep(time.Millisecond * 50)
	}
}

// main loop
func (app *App) update() {

}
