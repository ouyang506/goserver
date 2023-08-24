package redismgr

import (
	"context"
	"errors"
	"redisproxy/config"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	once               = sync.Once{}
	redisMgr *RedisMgr = nil
)

// singleton
func Instance() *RedisMgr {
	once.Do(func() {
		redisMgr = newRedisMgr()
	})
	return redisMgr
}

type RedisCmdClient interface {
	Do(ctx context.Context, args ...interface{}) *redis.Cmd
}

type RedisMgr struct {
	client RedisCmdClient
}

func newRedisMgr() *RedisMgr {
	mgr := &RedisMgr{}
	mgr.initClient()
	return mgr
}

func (mgr *RedisMgr) Start() {
}

func (mgr *RedisMgr) initClient() {
	conf := config.GetConfig()
	if conf.RedisConf.Type == 0 {
		mgr.client = redis.NewClient(&redis.Options{
			Addr:         conf.RedisConf.Endpoints.Items[0],
			Username:     conf.RedisConf.Username,
			Password:     conf.RedisConf.Password,
			PoolSize:     conf.RedisConf.PoolMaxConn,
			MaxIdleConns: conf.RedisConf.PoolMaxConn,
			MinIdleConns: conf.RedisConf.PoolMaxConn / 2,
		})
	} else if conf.RedisConf.Type == 1 {
		mgr.client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        conf.RedisConf.Endpoints.Items,
			Username:     conf.RedisConf.Username,
			Password:     conf.RedisConf.Password,
			PoolSize:     conf.RedisConf.PoolMaxConn,
			MaxIdleConns: conf.RedisConf.PoolMaxConn,
			MinIdleConns: conf.RedisConf.PoolMaxConn / 2,
		})
	}
}

func (mgr *RedisMgr) DoCmd(args ...any) (any, error) {
	if mgr.client == nil {
		return nil, errors.New("client is nil")
	}
	return mgr.client.Do(context.TODO(), args...).Result()
}

func (mgr *RedisMgr) IsNilError(err error) bool {
	return errors.Is(err, redis.Nil)
}
