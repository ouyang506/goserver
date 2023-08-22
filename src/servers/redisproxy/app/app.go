package app

import (
	"common/redisutil"
	"fmt"
	"framework/log"
	"os"
	"redisproxy/config"
	"redisproxy/netmgr"
	"redisproxy/redismgr"
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
	conf := config.GetConfig()
	if conf == nil {
		fmt.Printf("load config error, program exit!")
		os.Exit(1)
	}

	//init logger
	logger := log.NewCommonLogger()
	logger.SetLogLevel(log.LogLevelDebug)
	logger.AddSink(log.NewStdLogSink())
	logger.AddSink(log.NewFileLogSink("redisproxy", "../../../logs/", log.RotateByHour))
	logger.Start()
	log.SetLogger(logger)

	return true
}

func (app *App) Start() {
	log.Info("********************************************************")
	log.Info("**************     redisproxy server   *****************")
	log.Info("********************************************************")
	log.Info("App Start ..")
	// log app config info
	log.Info("App config info :")
	log.Info("%+v", config.GetConfig())

	redismgr.GetRedisMgr().Start()
	netmgr.GetNetMgr().Start()

	for {
		app.update()
		time.Sleep(time.Millisecond * 50 * 10000)
	}
}

// main loop
func (app *App) update() {
	redisutil.HSet("htest", "f1", "v1")

	v, err := redisutil.HGet("htest", "f12")
	log.Debug("get result, err: %v, value: %v", err, v)

	v2, err := redisutil.HMGet("htest", "f1", "f2")
	log.Debug("get result, err: %v, value: %v", err, v2)
}
