package rpc

import (
	"utility/consistent"
)

type RpcRouteType int

const (
	ConsistHashRoute RpcRouteType = iota
)

type RpcRouter interface {
	UpdateRoute(memberKey string, memberValue interface{})
	DelRoute(memberKey string)
	SelectRoute(key string) interface{}
}

func NewRpcRouter(routeType RpcRouteType) RpcRouter {
	var router RpcRouter = nil
	switch routeType {
	case ConsistHashRoute:
		{
			router = newConsistRouter()
		}
	}
	return router
}

type ConsistRouter struct {
	consistHash  *consistent.Consistent
	routeInfoMap map[string]interface{}
}

func newConsistRouter() *ConsistRouter {
	r := &ConsistRouter{
		consistHash:  consistent.NewConsistent(),
		routeInfoMap: map[string]interface{}{},
	}
	return r
}

func (r *ConsistRouter) UpdateRoute(memberKey string, memberValue interface{}) {
	r.consistHash.Add(memberKey)
	r.routeInfoMap[memberKey] = memberValue
}

func (r *ConsistRouter) DelRoute(memberKey string) {
	r.consistHash.Remove(memberKey)
	delete(r.routeInfoMap, memberKey)
}

func (r *ConsistRouter) SelectRoute(key string) interface{} {
	member := r.consistHash.Get(key)
	if member == "" {
		return nil
	}

	v, ok := r.routeInfoMap[member]
	if !ok {
		return nil
	}

	return v
}
