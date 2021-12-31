package registry

import (
	"common/log"
	"sync"
	"time"
)

type Registry interface {
	RegService(key string, value string, ttl uint32) error
	GetServices() (map[string]string, error)
	Watch() (chan WatchEvent, error)
}

type WatchCB func(WatchEventType, string, string)

type WatchEventType int

const (
	WatchEventTypeNone   WatchEventType = 0
	WatchEventTypeUpdate WatchEventType = 1
	WatchEventTypeDelete WatchEventType = 2
)

type WatchEvent struct {
	err       error
	eventType WatchEventType
	key       string
	value     string
}

type RegistryMgr struct {
	logger   log.Logger
	handler  Registry
	services sync.Map // map[string]string
}

func NewRegistryMgr(logger log.Logger, handler Registry) *RegistryMgr {

	mgr := &RegistryMgr{
		logger:  logger,
		handler: handler,
	}

	return mgr
}

func (mgr *RegistryMgr) DoRegister(key string, value string, ttl uint32) {
	//register self
	go mgr.regService(key, value, ttl)

	// watch service change event
	go mgr.watch()
}

func (mgr *RegistryMgr) regService(key, value string, ttl uint32) {
	for {
		err := mgr.handler.RegService(key, value, ttl)
		if err != nil {
			mgr.logger.LogError("RegService error : %s", err)
			time.Sleep(1 * time.Second)
			continue
		}
	}
}

func (mgr *RegistryMgr) watch() {
	for {
		watchChan, err := mgr.handler.Watch()
		if err != nil {
			mgr.logger.LogError("watch error : %s", err)
			time.Sleep(1 * time.Second)
			continue
		}

		serviceMap, err := mgr.handler.GetServices()
		if err != nil {
			mgr.logger.LogError("get services error : %s", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for k, v := range serviceMap {
			mgr.services.Store(k, v)
		}

		for event := range watchChan {
			if event.eventType == WatchEventTypeUpdate {
				mgr.services.Store(event.key, event.value)
			} else if event.eventType == WatchEventTypeDelete {
				mgr.services.Delete(event.key)
			}
		}

	}
}
