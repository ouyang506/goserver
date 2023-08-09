package registry

import (
	"fmt"
	"framework/log"
	"time"
)

const (
	RegistryDefaultTTL = 15
)

// service unique key
type ServiceKey struct {
	ServerType int
	IP         string
	Port       int
}

func (skey ServiceKey) String() string {
	return fmt.Sprintf("%d_%s_%d", skey.ServerType, skey.IP, skey.Port)
}

type Registry interface {
	RegService(key ServiceKey, ttl uint32) error
	GetServices() ([]ServiceKey, error)
	Watch() (chan WatchEvent, error)
}

type WatchEventType int

const (
	WatchEventTypeNone   WatchEventType = 0
	WatchEventTypeAdd    WatchEventType = 1
	WatchEventTypeDelete WatchEventType = 2
)

type WatchEvent struct {
	err       error
	eventType WatchEventType
	skey      ServiceKey
}

type OperType int

const (
	OperAdd    OperType = 1
	OperDelete OperType = 2
)

type HandleServiceCb func(oper OperType, skey ServiceKey)

type RegistryMgr struct {
	stub     Registry
	services map[ServiceKey]any
	cb       HandleServiceCb
}

func NewRegistryMgr(stub Registry, cb HandleServiceCb) *RegistryMgr {
	mgr := &RegistryMgr{
		stub:     stub,
		services: make(map[ServiceKey]any),
		cb:       cb,
	}
	return mgr
}

func (mgr *RegistryMgr) DoRegister(skey ServiceKey, ttl uint32) {
	//register self
	go mgr.regService(skey, ttl)

	// watch service change event
	go mgr.watch()
}

func (mgr *RegistryMgr) regService(skey ServiceKey, ttl uint32) {
	for {
		err := mgr.stub.RegService(skey, ttl)
		if err != nil {
			log.Error("RegService error : %s", err)
			delta := ttl / 2
			if delta < 1 {
				delta = 1
			}
			time.Sleep(time.Duration(delta) * time.Second)
			continue
		}
	}
}

func (mgr *RegistryMgr) watch() {
	for {
		mgr.clearAllServices()

		watchChan, err := mgr.stub.Watch()
		if err != nil {
			log.Error("watch error : %s", err)
			time.Sleep(1 * time.Second)
			continue
		}

		services, err := mgr.stub.GetServices()
		if err != nil {
			log.Error("get services error : %s", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for _, v := range services {
			mgr.services[v] = nil
		}

		for event := range watchChan {
			if event.eventType == WatchEventTypeAdd {
				mgr.onServiceAdd(event.skey)
			} else if event.eventType == WatchEventTypeDelete {
				mgr.onServiceDelete(event.skey)
			}
		}
	}
}

func (mgr *RegistryMgr) clearAllServices() {
	log.Info("registry manager clear all service : %+v", mgr.services)

	if len(mgr.services) <= 0 {
		return
	}

	for k := range mgr.services {
		mgr.cb(OperDelete, k)
	}

	mgr.services = map[ServiceKey]any{}
}

func (mgr *RegistryMgr) onServiceAdd(skey ServiceKey) {
	_, ok := mgr.services[skey]
	if !ok {
		log.Info("registry manager on add service %v", skey)
		mgr.services[skey] = nil
		mgr.cb(OperAdd, skey)
	}
}

func (mgr *RegistryMgr) onServiceDelete(skey ServiceKey) {
	_, ok := mgr.services[skey]
	if !ok {
		log.Info("registry manager on delete service not found, skey:%v", skey)
	} else {
		log.Info("registry manager on delete service %v", skey)
		delete(mgr.services, skey)
		mgr.cb(OperDelete, skey)
	}
}
