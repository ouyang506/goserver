package registry

import (
	"common/log"
	"time"
)

type Registry interface {
	RegService(key string, value string, ttl uint32) error
	GetServices() (map[string]string, error)
	Watch() (chan WatchEvent, error)
}

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

type OperType int

const (
	OperUpdate OperType = 1
	OperDelete OperType = 2
)

type HandleServiceCb func(OperType, string, string)

type RegistryMgr struct {
	logger   log.Logger
	stub     Registry
	services map[string]string
	cb       HandleServiceCb
}

func NewRegistryMgr(logger log.Logger, stub Registry, cb HandleServiceCb) *RegistryMgr {

	mgr := &RegistryMgr{
		logger:   logger,
		stub:     stub,
		services: make(map[string]string),
		cb:       cb,
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
		err := mgr.stub.RegService(key, value, ttl)
		if err != nil {
			mgr.logger.LogError("RegService error : %s", err)
			time.Sleep(1 * time.Second)
			continue
		}
	}
}

func (mgr *RegistryMgr) watch() {
	for {
		mgr.clearAllServices()

		watchChan, err := mgr.stub.Watch()
		if err != nil {
			mgr.logger.LogError("watch error : %s", err)
			time.Sleep(1 * time.Second)
			continue
		}

		serviceMap, err := mgr.stub.GetServices()
		if err != nil {
			mgr.logger.LogError("get services error : %s", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for k, v := range serviceMap {
			mgr.services[k] = v
		}

		for event := range watchChan {
			if event.eventType == WatchEventTypeUpdate {
				mgr.onServiceUpdate(event.key, event.value)
			} else if event.eventType == WatchEventTypeDelete {
				mgr.onServiceDelete(event.key)
			}
		}
	}
}

func (mgr *RegistryMgr) clearAllServices() {
	mgr.logger.LogDebug("registry manager clear all service : %+v", mgr.services)

	if len(mgr.services) <= 0 {
		return
	}

	for k, v := range mgr.services {
		mgr.cb(OperDelete, k, v)
	}

	mgr.services = map[string]string{}
}

func (mgr *RegistryMgr) onServiceUpdate(key, value string) {
	data, ok := mgr.services[key]
	if !ok {
		mgr.logger.LogDebug("registry manager on add service, key:%v, value:%v", key, value)
		mgr.services[key] = value
		mgr.cb(OperUpdate, key, value)
	} else {
		if data == value {
			//pass
		} else {
			mgr.logger.LogDebug("registry manager on update service, key:%v, value:%v", key, value)
			mgr.services[key] = value
			mgr.cb(OperUpdate, key, value)
		}
	}
}

func (mgr *RegistryMgr) onServiceDelete(key string) {
	data, ok := mgr.services[key]
	if !ok {
		//pass
	} else {
		mgr.logger.LogDebug("registry manager on delete service, key:%v", key)
		delete(mgr.services, key)
		mgr.cb(OperDelete, key, data)
	}
}
