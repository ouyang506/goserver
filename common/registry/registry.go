package registry

type Registry interface {
	DoRegister(key string, value string, ttl uint32)
}

type RegistryMgr struct {
	handler Registry
}

func NewRegistryMgr(handler Registry) *RegistryMgr {
	mgr := &RegistryMgr{}
	mgr.handler = handler

	return mgr
}

func (mgr *RegistryMgr) DoRegister(key string, value string, ttl uint32) {
	mgr.handler.DoRegister(key, value, ttl)
}
