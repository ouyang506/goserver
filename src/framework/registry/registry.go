package registry

const (
	RegistryDefaultTTL = 120
)

// service unique key
type ServiceKey struct {
	ServerType int
	IP         string
	Port       int
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
