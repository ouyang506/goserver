package registry

import (
	"context"
	"fmt"
	"framework/log"
	"strconv"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	EtcdBasePath = "/services/"
)

type EtcdRegistry struct {
	etcdCfg    *clientv3.Config
	etcdClient *clientv3.Client
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
			DialTimeout: 2 * time.Second,
			Username:    conf.Username,
			Password:    conf.Password,
		},
		etcdClient: nil,
	}

	return etcdRegistry
}

func (reg *EtcdRegistry) lazyInit() error {
	if reg.etcdClient == nil {
		cli, err := clientv3.New(*reg.etcdCfg)
		if err != nil {
			return err
		}
		reg.etcdClient = cli
	}
	return nil
}

// func (reg *EtcdRegistry) close() {
// 	if reg.etcdClient != nil {
// 		reg.etcdClient.Close()
// 		reg.etcdClient = nil
// 	}
// }

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

func (reg *EtcdRegistry) RegService(skey ServiceKey, ttl uint32) error {
	if err := reg.lazyInit(); err != nil {
		log.Error("init etcd client error : %s", err)
		return err
	}

	leaseCtx, leaseCancel := context.WithTimeout(context.Background(), time.Second*time.Duration(ttl))
	defer leaseCancel()
	lease := clientv3.NewLease(reg.etcdClient)
	leaseResp, err := lease.Grant(leaseCtx, int64(ttl))
	if err != nil {
		return err
	}

	leaseId := leaseResp.ID

	kv := clientv3.NewKV(reg.etcdClient)

	kvCtx, kvCancelFunc := context.WithTimeout(context.Background(), time.Duration(ttl*1000)*time.Millisecond)
	defer kvCancelFunc()
	_, err = kv.Put(kvCtx, EtcdBasePath+reg.makeKey(skey), "", clientv3.WithLease(leaseId))
	if err != nil {
		return err
	}

	tick := time.NewTicker(time.Duration(ttl*1000/10) * time.Millisecond)
	defer tick.Stop()
	for {
		<-tick.C
		if err := reg.renewLease(leaseId); err != nil {
			return err
		}
	}
}

func (reg *EtcdRegistry) renewLease(leaseId clientv3.LeaseID) error {
	if err := reg.lazyInit(); err != nil {
		log.Error("init etcd client error : %s", err)
		return err
	}

	leaseCtx, leaseCancel := context.WithTimeout(context.Background(), time.Duration(1000)*time.Millisecond)
	defer leaseCancel()
	lease := clientv3.NewLease(reg.etcdClient)
	_, err := lease.KeepAliveOnce(leaseCtx, leaseId)
	if err != nil {
		return err
	}
	return nil
}

func (reg *EtcdRegistry) GetServices() ([]ServiceKey, error) {
	if err := reg.lazyInit(); err != nil {
		log.Error("init etcd client error : %s", err)
		return nil, err
	}

	kvCtx, kvCancelFunc := context.WithTimeout(context.Background(), time.Duration(1000)*time.Millisecond)
	defer kvCancelFunc()

	kv := clientv3.NewKV(reg.etcdClient)
	resp, err := kv.Get(kvCtx, EtcdBasePath, clientv3.WithPrefix())
	if err != nil {
		return nil, err
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
	return ret, nil
}

func (reg *EtcdRegistry) Watch() (chan WatchEvent, error) {
	if err := reg.lazyInit(); err != nil {
		log.Error("init etcd client error : %s", err)
		return nil, err
	}

	watcher := clientv3.NewWatcher(reg.etcdClient)
	//defer watcher.Close()

	watchCtx := context.Background()
	watchChann := watcher.Watch(watchCtx, EtcdBasePath, clientv3.WithPrefix())

	retChan := make(chan WatchEvent)
	go func() {
		for {
			watchResp, ok := <-watchChann
			if !ok {
				retChan <- WatchEvent{
					err: fmt.Errorf("channel closed"),
				}
				close(retChan)
				return
			}

			if err := watchResp.Err(); err != nil {
				retChan <- WatchEvent{
					err: err,
				}
				close(retChan)
				return
			}

			if watchResp.Events == nil {
				continue
			}

			for _, event := range watchResp.Events {
				eventType := WatchEventTypeNone
				if event.Type == clientv3.EventTypePut {
					eventType = WatchEventTypeAdd
				} else if event.Type == clientv3.EventTypeDelete {
					eventType = WatchEventTypeDelete
				} else {
					continue //不存在此类情况
				}

				skey, err := reg.parseKey(string(event.Kv.Key))
				if err != nil {
					log.Error("parse service key error: %v, event type : %v", err, clientv3.EventTypePut)
				}

				retChan <- WatchEvent{
					err:       nil,
					eventType: eventType,
					skey:      skey,
				}
			}
		}
	}()

	return retChan, nil
}
