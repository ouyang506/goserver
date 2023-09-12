package app

import (
	"fmt"
	"framework/actor"
	"framework/log"
	"player/configmgr"

	"os"
	"player/netmgr"
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
	actorSystem *actor.ActorSystem
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
	logger.AddSink(log.NewStdLogSink())
	logger.AddSink(log.NewFileLogSink("gate", "../logs/", log.RotateByHour))
	logger.Start()
	log.SetLogger(logger)

	// init actor system
	app.actorSystem = actor.NewActorSystem()

	return true
}

func (app *App) Start() {
	log.Info("********************************************************")
	log.Info("*******************     player server   ****************")
	log.Info("********************************************************")
	log.Info("App Start ..")
	// log app config info
	log.Info("App config info :")
	log.Info("%+v", configmgr.Instance().GetConfig())

	root := app.actorSystem.Root()
	netmgr.NewNetMgr(root).Start()

	for {
		app.update()
		time.Sleep(time.Millisecond * 50)
	}
}

func (app *App) update() {

}
