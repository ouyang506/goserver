package app

import (
	"fmt"
	"framework/log"
	"login/configmgr"
	"login/handler"
	"login/netmgr"
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
	httpEngine *gin.Engine
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
	logger.AddSink(log.NewFileLogSink("login", "../logs/", log.RotateByHour))
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
	log.Info("%+v", configmgr.Instance().GetConfig())

	// start rpc net mgr
	netmgr.Instance().Start()

	// start http server
	conf := configmgr.Instance().GetConfig()
	addr := fmt.Sprintf("%s:%d", conf.HttpListen.IP, conf.HttpListen.Port)
	app.httpEngine.Run(addr)
}
