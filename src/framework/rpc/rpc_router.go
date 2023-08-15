package rpc

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
	"utility/consistent"

	"golang.org/x/exp/slices"
)

type RpcRouteType int

const (
	RandomRoute  RpcRouteType = 1 //随机路由
	ModRoute     RpcRouteType = 2 //同余路由
	ConsistRoute RpcRouteType = 3 //一致性路由
)

type RpcRouter interface {
	UpdateMember(memberKey string, memberValue any)
	DeleteMember(memberKey string)
	SelectMember(key any) any
}

func NewRpcRouter(routeType RpcRouteType) RpcRouter {
	var router RpcRouter = nil
	switch routeType {
	case RandomRoute:
		router = newRandomRouter()
	case ModRoute:
		router = newModRouter()
	case ConsistRoute:
		router = newConsistRouter()
	}
	return router
}

// 随机路由
type RandomRouter struct {
	mu        sync.RWMutex
	memberMap map[string]any
	memberArr []string
	rand      *rand.Rand
}

func newRandomRouter() *RandomRouter {
	return &RandomRouter{
		memberMap: make(map[string]any),
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *RandomRouter) UpdateMember(memberKey string, memberValue interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.memberMap[memberKey]; !ok {
		r.memberArr = append(r.memberArr, memberKey)
	}
	r.memberMap[memberKey] = memberValue
}

func (r *RandomRouter) DeleteMember(memberKey string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, v := range r.memberArr {
		if v == memberKey {
			r.memberArr = slices.Delete(r.memberArr, i, i+1)
			break
		}
	}
	delete(r.memberMap, memberKey)
}

func (r *RandomRouter) SelectMember(key any) any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	size := len(r.memberArr)
	if size <= 0 {
		return nil
	}

	i := int(r.rand.Int31()) % size
	memberKey := r.memberArr[i]

	v, ok := r.memberMap[memberKey]
	if !ok {
		return nil
	}
	return v
}

// 同余路由
type ModRouter struct {
	mu        sync.RWMutex
	memberMap map[string]any
	memberArr []string
}

func newModRouter() *ModRouter {
	return &ModRouter{
		memberMap: make(map[string]any),
	}
}

func (r *ModRouter) UpdateMember(memberKey string, memberValue interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	i, ok := slices.BinarySearch(r.memberArr, memberKey)
	if !ok {
		r.memberArr = slices.Insert(r.memberArr, i, memberKey)
	}

	r.memberMap[memberKey] = memberValue
}

func (r *ModRouter) DeleteMember(memberKey string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	i, ok := slices.BinarySearch(r.memberArr, memberKey)
	if ok {
		r.memberArr = slices.Delete(r.memberArr, i, i+1)
	}
	delete(r.memberMap, memberKey)
}

func (r *ModRouter) SelectMember(key any) any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	size := len(r.memberArr)
	if size <= 0 {
		return nil
	}

	value := uint64(0)
	switch tmp := key.(type) {
	case int:
		value = uint64(tmp)
	case int8:
		value = uint64(tmp)
	case int32:
		value = uint64(tmp)
	case int64:
		value = uint64(tmp)
	case uint:
		value = uint64(tmp)
	case uint8:
		value = uint64(tmp)
	case uint32:
		value = uint64(tmp)
	case uint64:
		value = uint64(tmp)
	default:
		i64, _ := strconv.ParseInt(fmt.Sprintf("%v", key), 10, 64)
		value = uint64(i64)
	}

	i := int(value % uint64(size))
	memberKey := r.memberArr[i]

	v, ok := r.memberMap[memberKey]
	if !ok {
		return nil
	}
	return v
}

// 一致性hash路由
type ConsistRouter struct {
	consistHash *consistent.Consistent
	members     map[string]any
	mu          sync.RWMutex
}

func newConsistRouter() *ConsistRouter {
	r := &ConsistRouter{
		consistHash: consistent.NewConsistent(),
		members:     map[string]any{},
	}
	return r
}

func (r *ConsistRouter) UpdateMember(memberKey string, memberValue any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.consistHash.Add(memberKey)
	r.members[memberKey] = memberValue
}

func (r *ConsistRouter) DeleteMember(memberKey string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.consistHash.Remove(memberKey)
	delete(r.members, memberKey)
}

func (r *ConsistRouter) SelectMember(key any) any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keyStr := ""
	switch tmp := key.(type) {
	case string:
		keyStr = tmp
	case []byte:
		keyStr = string(tmp)
	default:
		keyStr = fmt.Sprintf("%v", key)
	}

	member := r.consistHash.Get(keyStr)
	if member == "" {
		return nil
	}

	v, ok := r.members[member]
	if !ok {
		return nil
	}

	return v
}
