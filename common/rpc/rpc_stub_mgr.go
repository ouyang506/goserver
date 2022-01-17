package rpc

type IdToStubMap map[int]*RpcStub

type RpcStubManger struct {
	allStubs        IdToStubMap
	serverTypeStubs map[int]IdToStubMap
}

func NewRpcStubManager() *RpcStubManger {
	mgr := &RpcStubManger{}
	mgr.allStubs = make(IdToStubMap)
	mgr.serverTypeStubs = make(map[int]IdToStubMap)
	return mgr
}

func (mgr *RpcStubManger) AddStub(stub *RpcStub) {
	mgr.allStubs[stub.InstID] = stub
	typeStubs, ok := mgr.serverTypeStubs[stub.ServerType]
	if !ok {
		mgr.serverTypeStubs[stub.ServerType] = make(IdToStubMap)
		mgr.serverTypeStubs[stub.ServerType][stub.InstID] = stub
	} else {
		typeStubs[stub.InstID] = stub
	}
}

func (mgr *RpcStubManger) DelStub(stub *RpcStub) {
	delete(mgr.allStubs, stub.InstID)
	typeStubs, ok := mgr.serverTypeStubs[stub.ServerType]
	if ok {
		delete(typeStubs, stub.InstID)
	}
}

func (mgr *RpcStubManger) GetStub(serverType int) *RpcStub {
	typeStubs, ok := mgr.serverTypeStubs[serverType]
	if !ok {
		return nil
	}
	if len(typeStubs) <= 0 {
		return nil
	}

	for _, v := range typeStubs {
		return v
	}

	return nil
}
