package keepalive

import (
	"common/redisutil"
	"fmt"
	"framework/actor"
	"framework/log"
	"framework/registry"
	"gate/configmgr"
	"strconv"
	"time"
)

const ActorName = "keep_alive_actor"

type KeepAliveActor struct {
	ttl   int
	timer *time.Timer
}

type TickReq struct{}
type TickResp struct{}

func NewKeepAliveActor() *KeepAliveActor {
	return &KeepAliveActor{}
}

func (ka *KeepAliveActor) Receive(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Start:
		ka.startTimer(ctx)

	case *TickReq:
		ka.regOuterAddrToRedis()

	case *actor.Stop:
		ka.stopTimer()
	}
}

func (ka *KeepAliveActor) startTimer(ctx actor.Context) {
	ka.ttl = registry.RegistryDefaultTTL
	delta := ka.ttl/2 - 1
	if delta < 1 {
		delta = 1
	}
	ctx.Send(ctx.Self(), &TickReq{}) //先发送一次
	ka.timer = time.AfterFunc(time.Duration(delta)*time.Second, func() {
		ctx.Send(ctx.Self(), &TickReq{})
		ka.timer.Reset(time.Duration(delta) * time.Second)
	})
}

func (ka *KeepAliveActor) stopTimer() {
	ka.timer.Stop()
}

// 将对外ip+port注册到redis，供login服务使用
// 注册中心只有对内的ip+port
func (ka *KeepAliveActor) regOuterAddrToRedis() {
	conf := configmgr.Instance().GetConfig()
	ip, port := conf.Outer.OuterIp, conf.Outer.Port
	addr := fmt.Sprintf("%v_%d", ip, port)

	expired := time.Now().Unix() + int64(ka.ttl)
	err := redisutil.HSet(redisutil.RKeyGateOuterAddr, addr, strconv.FormatInt(expired, 10))
	if err != nil {
		log.Error("register gate outer addr to redis error: %v, gate addr: %v", err, addr)
	}
}
