package registry

import (
	"common/log"
	"time"
)

type Registry interface {
	RegService(key string, value string, ttl uint32) error
	GetServices() (map[string]string, error)
	Watch(WatchCB) error
}

type WatchCB func(RegistryEventType, string, string)

type RegistryEventType int

const (
	RegistryEventNone   RegistryEventType = 0
	RegistryEventUpdate RegistryEventType = 1
	RegistryEventDelete RegistryEventType = 2
)

type RegistryMgr struct {
	logger  log.Logger
	handler Registry
	// key     string
	// value   string
	// ttl     uint32
}

func NewRegistryMgr(logger log.Logger, handler Registry) *RegistryMgr {

	mgr := &RegistryMgr{
		logger:  logger,
		handler: handler,
	}

	return mgr
}

func (mgr *RegistryMgr) DoRegister(key string, value string, ttl uint32) {
	go func() {
		for {
			err := mgr.handler.RegService(key, value, ttl)
			if err != nil {
				mgr.logger.LogError("etcd run registry error : %s", err)
				time.Sleep(1 * time.Second)
			}
		}
	}()

	go func() {
		for {
			err := mgr.handler.Watch(mgr.watchServiceCb)
			if err != nil {
				mgr.logger.LogError("etcd run registry error : %s", err)
				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func (mgr *RegistryMgr) watchServiceCb(eventType RegistryEventType, key, value string) {

}
