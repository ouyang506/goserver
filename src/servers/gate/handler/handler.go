package handler

import (
	"framework/proto/pb"
	"reflect"
)

type MessageHandler struct {
}

func (h *MessageHandler) GetMsgHandlerByName(msgname string) *reflect.Value {
	//refType := reflect.TypeOf(*h)
	refValue := reflect.ValueOf(h)

	methodName := "Handle" + msgname
	method := refValue.MethodByName(methodName)
	if method.Kind() != reflect.Func {
		return nil
	}
	return &method
}

func (h *MessageHandler) GetMsgHandlerById(msgId int) (*reflect.Value, reflect.Value, reflect.Value) {
	for _, rpcInfo := range pb.CsMsgArray {
		if rpcInfo[1].(int) == msgId {
			reqRefType := reflect.TypeOf(rpcInfo[2])
			//reqRefValue :=  reqRefType.Elem().Name()
			reqMsgName := reqRefType.Elem().Name()

			retHandler := h.GetMsgHandlerByName(reqMsgName)

			req := reflect.New(reflect.TypeOf(rpcInfo[2]).Elem())
			resp := reflect.New(reflect.TypeOf(rpcInfo[3]).Elem())

			//in := []reflect.Value{req, resp}
			return retHandler, req, resp
		}
	}
	return nil, reflect.Value{}, reflect.Value{}
}
