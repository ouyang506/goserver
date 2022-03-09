package rpc

type RpcManager struct {
	rpcStubMgr *RpcStubManger
	rpcMap     map[int64]*Rpc
}

func NewRpcManager() *RpcManager {
	mgr := &RpcManager{}
	mgr.rpcStubMgr = NewRpcStubManager()
	return mgr
}

func (mgr *RpcManager) GetStubMgr() *RpcStubManger {
	return mgr.rpcStubMgr
}

func (mgr *RpcManager) AddRpc(rpc *Rpc) bool {
	_, ok := mgr.rpcMap[rpc.SessionID]
	if ok {
		return false
	}

	rpcStub := mgr.rpcStubMgr.FindStub(rpc)
	if rpcStub == nil {
		return false
	}

	mgr.rpcMap[rpc.SessionID] = rpc

	return true
}

func (mgr *RpcManager) RemoveRpc(rpc *Rpc) bool {
	_, ok := mgr.rpcMap[rpc.SessionID]
	if ok {
		return false
	}
	delete(mgr.rpcMap, rpc.SessionID)
	return true
}
