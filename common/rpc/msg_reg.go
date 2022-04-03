package rpc

import (
	"common/pbmsg"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

//type MsgHandler func(req proto.Message, resp proto.Message)
type MsgHandler struct {
	method reflect.Method
}

var (
	allMsg        map[int]protoreflect.Message // msgId -> msg反射类
	msgIdMap      map[int]int                  // 请求msgID->返回msgID
	msgHandlerMap map[int]reflect.Value        // 请求msgID->处理函数
)

// 程序启动时注册，非线程安全
func InitMsgMapping() {
	RegProtoMsgMapping(pbmsg.MsgID_login_gate_req, &pbmsg.LoginGateReqT{},
		pbmsg.MsgID_login_gate_resp, &pbmsg.LoginGateRespT{})
}

// 程序启动时注册，非线程安全
func RegMsgHandler(reqMsgID pbmsg.MsgID, handleFunc interface{}) {
	msgHandlerMap[int(reqMsgID)] = reflect.ValueOf(handleFunc)
}

// 注册协议映射和处理函数
func RegProtoMsgMapping(reqMsgId pbmsg.MsgID, reqMsg proto.Message, respMsgId pbmsg.MsgID, respMsg proto.Message) {
	var reqRelect protoreflect.Message = nil
	if reqMsgId > 0 && reqMsg != nil {
		reqRelect = reqMsg.ProtoReflect()
		if !reqRelect.IsValid() {
			reqRelect = nil
		}
	}

	var respRelect protoreflect.Message = nil
	if respMsgId > 0 && respMsg != nil {
		respRelect = respMsg.ProtoReflect()
		if !respRelect.IsValid() {
			respRelect = nil
		}
	}

	allMsg[int(reqMsgId)] = reqRelect
	allMsg[int(respMsgId)] = respRelect

	if reqMsgId > 0 && respMsgId > 0 {
		msgIdMap[int(reqMsgId)] = int(respMsgId)
	}
}

// 通过msgID创建msg
// TODO : 此处需要一个pool缓存
func GetProtoMsgById(msgId int) proto.Message {
	msgRelect, ok := allMsg[msgId]
	if !ok {
		return nil
	}
	dst := msgRelect.New()
	return dst.Interface()
}

// 通过请求msgID获取处理函数
func GetProtoMsgHandler(msgId int) reflect.Value {
	handler, ok := msgHandlerMap[msgId]
	if !ok {
		return reflect.Value{}
	}
	return handler
}
