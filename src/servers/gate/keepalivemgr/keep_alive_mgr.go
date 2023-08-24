package keepalivemgr

import (
	"common/redisutil"
	"context"
	"fmt"
	"framework/log"
	"framework/registry"
	"gate/configmgr"
	"strconv"
	"sync"
	"time"
)

var (
	once                       = sync.Once{}
	keepAliveMgr *KeepAliveMgr = nil
)

// singleton
func Instance() *KeepAliveMgr {
	once.Do(func() {
		keepAliveMgr = newKeepAliveMgr()
	})
	return keepAliveMgr
}

type KeepAliveMgr struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func newKeepAliveMgr() *KeepAliveMgr {
	ctx, ctxCancel := context.WithCancel(context.TODO())
	mgr := &KeepAliveMgr{
		ctx:       ctx,
		ctxCancel: ctxCancel,
	}
	return mgr
}

func (mgr *KeepAliveMgr) Start() {
	go mgr.RegOuterAddrToRedis()
}

func (mgr *KeepAliveMgr) Stop() {
	mgr.ctxCancel()
}

// 将对外ip+port注册到redis，供login服务使用
func (mgr *KeepAliveMgr) RegOuterAddrToRedis() {
	conf := configmgr.Instance().GetConfig()
	ip, port := conf.Outer.OuterIp, conf.Outer.Port
	addr := fmt.Sprintf("%v_%d", ip, port)

	ttl := registry.RegistryDefaultTTL
	delta := ttl/2 - 1
	if delta < 1 {
		delta = 1
	}
	ticker := time.NewTicker(time.Duration(delta) * time.Second)
	defer ticker.Stop()
	stop := false
	for {
		select {
		case <-ticker.C:
			expired := time.Now().Unix() + int64(ttl)
			err := redisutil.HSet(redisutil.RKeyGateOuterAddr, addr, strconv.FormatInt(expired, 10))
			if err != nil {
				log.Error("register gate outer addr to redis error: %v, gate addr: %v", err, addr)
			}
		case <-mgr.ctx.Done():
			stop = true
		}
		if stop {
			break
		}
	}
}
