package registry

import (
	"context"
	"fmt"
	"framework/log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	EtcdBasePath = "/services"
)

type EtcdRegistry struct {
	logger     log.Logger
	etcdCfg    *clientv3.Config
	etcdClient *clientv3.Client
}

func NewEtcdRegistry(logger log.Logger, endpoints []string,
	username string, password string) *EtcdRegistry {

	etcdRegistry := &EtcdRegistry{
		logger: logger,
		etcdCfg: &clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: 2 * time.Second,
			Username:    username,
			Password:    password,
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

func (reg *EtcdRegistry) RegService(key string, value string, ttl uint32) error {
	if err := reg.lazyInit(); err != nil {
		reg.logger.LogError("init etcd client error : %s", err)
		return err
	}

	leaseCtx, leaseCancel := context.WithTimeout(context.Background(), time.Duration(ttl*1000)*time.Millisecond)
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
	_, err = kv.Put(kvCtx, EtcdBasePath+key, value, clientv3.WithLease(leaseId))
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
		reg.logger.LogError("init etcd client error : %s", err)
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

func (reg *EtcdRegistry) GetServices() (map[string]string, error) {
	if err := reg.lazyInit(); err != nil {
		reg.logger.LogError("init etcd client error : %s", err)
		return nil, err
	}

	kvCtx, kvCancelFunc := context.WithTimeout(context.Background(), time.Duration(1000)*time.Millisecond)
	defer kvCancelFunc()

	kv := clientv3.NewKV(reg.etcdClient)
	resp, err := kv.Get(kvCtx, EtcdBasePath, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	ret := make(map[string]string)
	if resp.Kvs != nil {
		for _, entry := range resp.Kvs {
			ret[string(entry.Key)] = string(entry.Value)
		}
	}

	return ret, nil
}

func (reg *EtcdRegistry) Watch() (chan WatchEvent, error) {
	if err := reg.lazyInit(); err != nil {
		reg.logger.LogError("init etcd client error : %s", err)
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
				key := string(event.Kv.Key)
				value := string(event.Kv.Value)
				eventType := WatchEventTypeNone
				if event.Type == clientv3.EventTypePut {
					eventType = WatchEventTypeUpdate
				} else if event.Type == clientv3.EventTypeDelete {
					eventType = WatchEventTypeDelete
				}
				retChan <- WatchEvent{
					err:       nil,
					eventType: eventType,
					key:       key,
					value:     value,
				}
			}
		}
	}()

	return retChan, nil
}
