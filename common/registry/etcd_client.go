package registry

import (
	"common/log"
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdRegistry struct {
	logger     log.Logger
	etcdCfg    *clientv3.Config
	etcdClient *clientv3.Client
}

func NewEtcdRegistry(logger log.Logger, endpoints []string,
	username string, password string,
	dailTimeout time.Duration) *EtcdRegistry {

	etcdRegistry := &EtcdRegistry{
		logger: logger,
		etcdCfg: &clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: dailTimeout,
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

func (reg *EtcdRegistry) close() {
	if reg.etcdClient != nil {
		reg.etcdClient.Close()
		reg.etcdClient = nil
	}
}

func (reg *EtcdRegistry) DoRegister(key string, value string, ttl uint32) {
	go func() {
		for {
			err := reg.doRegister(key, value, ttl)
			if err != nil {
				reg.logger.LogError("etcd run registry error : %s", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for {
			err := reg.watch("/services")
			if err != nil {
				reg.logger.LogError("etcd run watch error : %s", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func (reg *EtcdRegistry) doRegister(key string, value string, ttl uint32) error {
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
	_, err = kv.Put(kvCtx, key, value, clientv3.WithLease(leaseId))
	if err != nil {
		return err
	}

	tick := time.NewTicker(time.Duration(ttl*1000/10) * time.Millisecond)
	defer tick.Stop()
	for {
		err = nil
		select {
		case <-tick.C:
			err = reg.renewLease(leaseId)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (reg *EtcdRegistry) renewLease(leaseId clientv3.LeaseID) error {
	leaseCtx, leaseCancel := context.WithTimeout(context.Background(), time.Duration(1000)*time.Millisecond)
	defer leaseCancel()
	lease := clientv3.NewLease(reg.etcdClient)
	_, err := lease.KeepAliveOnce(leaseCtx, leaseId)
	if err != nil {
		return err
	}
	return nil
}

func (reg *EtcdRegistry) getRegistryInfo(prefix string) (map[string]string, error) {
	kvCtx, kvCancelFunc := context.WithTimeout(context.Background(), time.Duration(1000)*time.Millisecond)
	defer kvCancelFunc()

	kv := clientv3.NewKV(reg.etcdClient)
	resp, err := kv.Get(kvCtx, prefix, clientv3.WithPrefix())
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

func (reg *EtcdRegistry) watch(prefix string) error {
	watcher := clientv3.NewWatcher(reg.etcdClient)
	defer watcher.Close()

	watchCtx := context.TODO()
	watchChann := watcher.Watch(watchCtx, prefix, clientv3.WithPrefix())

	for {
		watchResp, ok := <-watchChann
		if !ok {
			return fmt.Errorf("channel closed")
		}

		if err := watchResp.Err(); err != nil {
			return err
		}

		if watchResp.Events == nil {
			continue
		}

		for _, event := range watchResp.Events {
			key := string(event.Kv.Key)
			value := string(event.Kv.Value)
			if event.Type == clientv3.EventTypePut {
				reg.logger.LogInfo("etcd watch put key : %v, value : %v", key, value)
			} else if event.Type == clientv3.EventTypeDelete {
				reg.logger.LogInfo("etcd watch delete key : %v, value : %v", key, value)
			}
		}
	}
	return nil
}
