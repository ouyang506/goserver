package registry

const (
	RegistryDefaultTTL = 15
)

// service unique key
type ServiceKey struct {
	ServerType int
	IP         string
	Port       int
}

func ServiceKeyCmp(a ServiceKey, b ServiceKey) int {
	if a == b {
		return 0
	}

	switch {
	case a.ServerType < b.ServerType:
		return -1
	case a.ServerType > b.ServerType:
		return 1
	}

	switch {
	case a.IP < b.IP:
		return -1
	case a.IP > b.IP:
		return 1
	}

	switch {
	case a.Port < b.Port:
		return -1
	case a.Port > b.Port:
		return 1
	}
	return 0
}

type WatchEventType int

const (
	WatchEventTypeAdd    WatchEventType = 1
	WatchEventTypeDelete WatchEventType = 2
)

type WatchCallback func(WatchEventType, ServiceKey)

// 服务注册
type Registry interface {
	// 注册自己的监听地址到注册中心
	RegService(key ServiceKey)
	// 拉取当前所有注册的服务，遍历调用callback(类型为ADD), 并且开始监听之后的事件
	FetchAndWatchService(WatchCallback)
	// 释放资源
	Close()
}
