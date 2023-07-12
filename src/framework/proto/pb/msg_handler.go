package pb

import "reflect"

// type MsgHandlerBase struct {
// 	MsgId2MethodMap map[int]any
// }

type MsgHandler interface {
}

func ReflectHandlerMethods(h MsgHandler) map[string]reflect.Value {
	ret := make(map[string]reflect.Value)
	refType := reflect.TypeOf(h)
	refValue := reflect.ValueOf(h)
	methodCount := refType.NumMethod()
	for i := 0; i < methodCount; i++ {
		methodName := refType.Method(i).Name
		_ = methodName
		ret[methodName] = refValue.Method(i)
	}
	return ret
}
