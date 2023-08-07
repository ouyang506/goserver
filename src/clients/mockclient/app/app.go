package app

import (
	"fmt"
	"framework/log"
	"mockclient/config"
	"mockclient/robot"
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
	conf  *config.Config
	robot *robot.Robot
}

func (app *App) init() bool {
	// init config
	app.conf = config.NewConfig()
	err := app.conf.Load("../../../conf/client.xml")
	if err != nil {
		fmt.Printf("Load config error, program exit! error message : %v", err)
		os.Exit(1)
	}

	//init logger
	logger := log.NewCommonLogger()
	logger.AddSink(log.NewStdLogSink())
	logger.AddSink(log.NewFileLogSink("mockclient", "../../../logs/", log.RotateByHour))
	logger.Start()
	log.SetLogger(logger)

	// init robot
	app.robot = robot.NewRobot(app.conf)

	return true
}

func (app *App) Start() {
	log.Info("====================================================================")
	log.Info("")
	log.Info("App start ..")

	// log app config info
	log.Info("App config info :")
	log.Info("%+v", app.conf)

	//app.netMgr.Start()

	for {
		app.update()
		time.Sleep(20 * time.Millisecond)
	}
}

func (app *App) update() {
	app.robot.Update()
}
