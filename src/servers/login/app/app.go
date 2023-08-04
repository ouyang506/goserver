package app

import (
	"fmt"
	"framework/log"
	"login/config"
	"login/handler"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
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
	conf       *config.Config
	httpEngine *gin.Engine
}

func (app *App) init() bool {
	// init config
	app.conf = config.NewConfig()
	err := app.conf.Load("../../../conf/login.xml")
	if err != nil {
		fmt.Printf("Load config error, program exit! error message : %v", err)
		os.Exit(1)
	}

	//init logger
	logger := log.NewCommonLogger()
	logger.AddSink(log.NewStdLogSink())
	logger.AddSink(log.NewFileLogSink("login", "../../../logs/", log.RotateByHour))
	logger.Start()
	log.SetLogger(logger)

	// init http server
	app.httpEngine = gin.Default()
	gin.SetMode(gin.DebugMode)
	//gin.SetMode(gin.ReleaseMode)
	app.httpEngine.SetTrustedProxies(nil)
	handler.RegHttpHandler(app.httpEngine)

	return true
}

func (app *App) Start() {
	log.Info("********************************************************")
	log.Info("********************     gate server   *****************")
	log.Info("********************************************************")
	log.Info("App Start ..")
	// log app config info
	log.Info("App config info :")
	log.Info("%+v", app.conf)

	addr := fmt.Sprintf("%s:%d", app.conf.HttpListen.IP, app.conf.HttpListen.Port)
	app.httpEngine.Run(addr)
}
