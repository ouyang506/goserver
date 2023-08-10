package app

import (
	"common"
	"fmt"
	"framework/log"
	"mysqlproxy/config"
	"mysqlproxy/dbmgr"
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
	conf := config.GetConfig()
	if conf == nil {
		fmt.Printf("load config error, program exit!")
		os.Exit(1)
	}

	//init logger
	logger := log.NewCommonLogger()
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
	log.Info("%+v", config.GetConfig())

	dbmgr.GetMysqlMgr().Start()
	netmgr.GetNetMgr().Start()

	for {
		app.update()
		time.Sleep(time.Millisecond * 20 * 10000)
	}
}

// main loop
func (app *App) update() {
	time.Sleep(5 * time.Second)
	result, err := common.QuerySQL("select * from account where username = ?", "admin")
	log.Debug("result: %+v, err: %v", result, err)
	// if err != nil {
	// 	for _, row := range result.Rows {
	// 		for _, field := range row {
	// 			log.Debug("%v", field.String())
	// 		}
	// 	}
	// }
}
