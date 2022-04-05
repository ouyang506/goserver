package rpc

import (
	"common/log"
	"common/pbmsg"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	allMsg        map[int]protoreflect.Message // msgId -> msg反射类
	msgIdMap      map[int]int                  // 请求msgID->返回msgID
	msgHandlerMap map[int]*reflect.Value       // 请求msgID->处理函数
)

// 程序启动时注册，非线程安全
func InitMsgMapping() {
	RegProtoMsgMapping(pbmsg.MsgID_login_gate_req, &pbmsg.LoginGateReqT{},
		pbmsg.MsgID_login_gate_resp, &pbmsg.LoginGateRespT{})
}

// 程序启动时注册，非线程安全
func RegMsgHandler(reqMsgID pbmsg.MsgID, handleFunc interface{}) {
	f := reflect.ValueOf(handleFunc)
	if f.Kind() != reflect.Func {
		log.Error("handleFunc is not a function type.")
		return
	}

	if msgHandlerMap == nil {
		msgHandlerMap = make(map[int]*reflect.Value)
	}
	msgHandlerMap[int(reqMsgID)] = &f
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

	if allMsg == nil {
		allMsg = make(map[int]protoreflect.Message)
	}
	allMsg[int(reqMsgId)] = reqRelect
	allMsg[int(respMsgId)] = respRelect

	if msgIdMap == nil {
		msgIdMap = make(map[int]int)
	}
	if reqMsgId > 0 && respMsgId > 0 {
		msgIdMap[int(reqMsgId)] = int(respMsgId)
	}
}

// 通过msgID创建msg
// TODO : 此处需要一个pool缓存
func GetProtoMsgById(msgId int) proto.Message {
	if allMsg == nil {
		return nil
	}

	msgRelect, ok := allMsg[msgId]
	if !ok {
		return nil
	}
	dst := msgRelect.New()
	return dst.Interface()
}

// 通过请求msgID获取处理函数
func GetProtoMsgHandler(msgId int) *reflect.Value {
	if msgHandlerMap == nil {
		return nil
	}

	handler, ok := msgHandlerMap[msgId]
	if !ok {
		return nil
	}
	return handler
}
