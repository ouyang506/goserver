package registry

import (
	"context"
	"fmt"
	"framework/log"
	"strconv"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/exp/slices"
)

const (
	EtcdBasePath = "/services/"
)

// Registry Interface Implementment
type EtcdRegistry struct {
	etcdCfg *clientv3.Config

	etcdClient   *clientv3.Client
	etcdClientMu sync.Mutex

	services   []ServiceKey
	servicesMu sync.Mutex
}

type EtcdConfig struct {
	Endpoints []string
	Username  string
	Password  string
}

func NewEtcdRegistry(conf EtcdConfig) *EtcdRegistry {
	etcdRegistry := &EtcdRegistry{
		etcdCfg: &clientv3.Config{
			Endpoints:   conf.Endpoints,
			DialTimeout: 5 * time.Second,
			Username:    conf.Username,
			Password:    conf.Password,
		},
	}

	return etcdRegistry
}

func (reg *EtcdRegistry) RegService(skey ServiceKey) {
	go func() {
		for {
			err := reg.doRegService(skey, RegistryDefaultTTL)
			if err != nil {
				log.Error("doRegService return error, %s", err)
				time.Sleep(time.Second * 1)
				continue
			}
		}
	}()
}

func (reg *EtcdRegistry) FetchAndWatchService(cb WatchCallback) {
	go func() {
		var revision int64
		var services []ServiceKey
		var err error

		for {
			services, revision, err = reg.getServices()
			if err != nil {
				log.Error("get services error : %s", err)
				time.Sleep(1 * time.Second)
				continue
			}

			retAdd, retDelete := reg.replaceAllServices(services)
			for _, skey := range retAdd {
				cb(WatchEventTypeAdd, skey)
			}
			for _, skey := range retDelete {
				cb(WatchEventTypeDelete, skey)
			}
			break
		}

		//watch events after the revison
		revision += 1

		for {
			err := reg.doWatch(&revision, cb)
			if err != nil {
				log.Error("do watch error, %v", err)
				continue
			}
		}
	}()
}

func (reg *EtcdRegistry) Close() {
	reg.etcdClientMu.Lock()
	defer reg.etcdClientMu.Unlock()

	if reg.etcdClient != nil {
		reg.etcdClient.Close()
		reg.etcdClient = nil
	}
}

func (reg *EtcdRegistry) lazyInit() error {
	reg.etcdClientMu.Lock()
	defer reg.etcdClientMu.Unlock()

	if reg.etcdClient == nil {
		cli, err := clientv3.New(*reg.etcdCfg)
		if err != nil {
			return err
		}
		reg.etcdClient = cli
	}
	return nil
}

func (reg *EtcdRegistry) addService(skey ServiceKey) bool {
	reg.servicesMu.Lock()
	defer reg.servicesMu.Unlock()

	index, ok := slices.BinarySearchFunc(reg.services, skey, ServiceKeyCmp)
	if ok {
		return false
	}

	reg.services = slices.Insert(reg.services, index, skey)
	return true
}

func (reg *EtcdRegistry) delService(skey ServiceKey) bool {
	reg.servicesMu.Lock()
	defer reg.servicesMu.Unlock()

	index := slices.Index(reg.services, skey)
	if index < 0 {
		return false
	}
	reg.services = slices.Delete(reg.services, index, index+1)
	return true
}

func (reg *EtcdRegistry) replaceAllServices(newServices []ServiceKey) (retAdd []ServiceKey, retDelete []ServiceKey) {
	reg.servicesMu.Lock()
	defer reg.servicesMu.Unlock()

	retAdd = []ServiceKey{}
	retDelete = []ServiceKey{}

	for _, skey := range newServices {
		_, ok := slices.BinarySearchFunc(reg.services, skey, ServiceKeyCmp)
		if ok {
			continue
		}
		retAdd = append(retAdd, skey)
	}

	for _, skey := range reg.services {
		if !slices.Contains(newServices, skey) {
			retDelete = append(retDelete, skey)
		}
	}

	reg.services = newServices
	return
}

func (reg *EtcdRegistry) doRegService(skey ServiceKey, ttl uint32) error {
	if err := reg.lazyInit(); err != nil {
		log.Error("init etcd client error : %s", err)
		return err
	}

	leaseCtx, leaseCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer leaseCancel()
	lease := clientv3.NewLease(reg.etcdClient)
	leaseResp, err := lease.Grant(leaseCtx, int64(ttl))
	if err != nil {
		return err
	}
	//不收回，避免续约失败后的服务间网络波动
	//defer lease.Revoke(context.Background(), leaseResp.ID)

	leaseId := leaseResp.ID
	kv := clientv3.NewKV(reg.etcdClient)

	kvCtx, kvCancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
	defer kvCancelFunc()
	_, err = kv.Put(kvCtx, EtcdBasePath+reg.makeKey(skey), "", clientv3.WithLease(leaseId))
	if err != nil {
		return err
	}

	aliveRespChan, err := lease.KeepAlive(context.Background(), leaseId)
	if err != nil {
		return err
	}

	for {
		_, ok := <-aliveRespChan
		if !ok {
			return fmt.Errorf("keepalive channel closed")
		}
	}
}

func (reg *EtcdRegistry) getServices() ([]ServiceKey, int64, error) {
	if err := reg.lazyInit(); err != nil {
		log.Error("init etcd client error : %s", err)
		return nil, 0, err
	}

	kvCtx, kvCancelFunc := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer kvCancelFunc()

	kv := clientv3.NewKV(reg.etcdClient)
	resp, err := kv.Get(kvCtx, EtcdBasePath, clientv3.WithPrefix())
	if err != nil {
		return nil, 0, err
	}

	skeyMap := make(map[ServiceKey]any)
	ret := []ServiceKey{}
	if resp.Kvs != nil {
		for _, entry := range resp.Kvs {
			skey, err := reg.parseKey(string(entry.Key))
			if err != nil {
				log.Error("parse service key error: %v, key: %v", err, string(entry.Key))
				continue
			}
			if _, ok := skeyMap[skey]; !ok {
				ret = append(ret, skey)
			}
			skeyMap[skey] = nil
		}
	}
	revision := resp.Header.Revision
	return ret, revision, nil
}

func (reg *EtcdRegistry) doWatch(revision *int64, cb WatchCallback) error {
	if err := reg.lazyInit(); err != nil {
		log.Error("init etcd client error : %s", err)
		return err
	}

	watcher := clientv3.NewWatcher(reg.etcdClient)
	defer watcher.Close()

	watchChann := watcher.Watch(context.Background(), EtcdBasePath,
		clientv3.WithPrefix(), clientv3.WithRev(*revision))

	for {
		watchResp, ok := <-watchChann
		if !ok {
			return fmt.Errorf("watch channel closed")
		}

		if err := watchResp.Err(); err != nil {
			return err
		}

		// assign the new revision for the cycle watch
		*revision = watchResp.Header.Revision

		for _, event := range watchResp.Events {
			var eventType WatchEventType
			switch event.Type {
			case clientv3.EventTypePut:
				eventType = WatchEventTypeAdd
			case clientv3.EventTypeDelete:
				eventType = WatchEventTypeDelete
			default:
				log.Error("invalid event type : %v", event.Type)
				continue
			}
			// log.Debug("Resp.Header.Revision = %v, create_version = %v , mod_version = %v ,eventType = %v, key = %v, value = %v",
			// 	watchResp.Header.Revision, event.Kv.CreateRevision, event.Kv.ModRevision, eventType,
			// 	string(event.Kv.Key), string(event.Kv.Value))

			skey, err := reg.parseKey(string(event.Kv.Key))
			if err != nil {
				log.Error("parse service key error: %v, event key : %v", err, string(event.Kv.Key))
				continue
			}

			if eventType == WatchEventTypeAdd {
				if !reg.addService(skey) {
					continue
				}
			} else if eventType == WatchEventTypeDelete {
				if !reg.delService(skey) {
					continue
				}
			}
			cb(eventType, skey)
		}
	}
}

func (reg *EtcdRegistry) makeKey(skey ServiceKey) string {
	return fmt.Sprintf("%d_%s_%d", skey.ServerType, skey.IP, skey.Port)
}

func (reg *EtcdRegistry) parseKey(key string) (ServiceKey, error) {
	key = strings.TrimPrefix(key, EtcdBasePath)
	skey := ServiceKey{}
	splits := strings.Split(key, "_")
	if len(splits) == 3 {
		skey.ServerType, _ = strconv.Atoi(splits[0])
		skey.IP = splits[1]
		skey.Port, _ = strconv.Atoi(splits[2])
	}
	return skey, nil
}
