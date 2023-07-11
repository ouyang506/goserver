package pb

import (
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	MsgId2Type map[int][]protoreflect.Message = make(map[int][]protoreflect.Message)
	MsgName2Id map[string]int                 = make(map[string]int)
)

func init() {
	fillMsgMap()
}

func fillMsgMap() {
	for _, info := range CsMsgArray {
		msgId := info[0].(int)
		arr := []protoreflect.Message{}
		arr = append(arr, info[1].(proto.Message).ProtoReflect())
		if len(info) >= 3 && info[2] != nil {
			arr = append(arr, info[1].(proto.Message).ProtoReflect())
		}
		MsgId2Type[msgId] = arr

		msgName := reflect.TypeOf(info[1]).Elem().Name()
		MsgName2Id[msgName] = msgId
	}

	for _, info := range SsMsgArray {
		msgId := info[0].(int)
		arr := []protoreflect.Message{}
		arr = append(arr, info[1].(proto.Message).ProtoReflect())
		if len(info) >= 3 && info[2] != nil {
			arr = append(arr, info[1].(proto.Message).ProtoReflect())
		}
		MsgId2Type[msgId] = arr

		msgName := reflect.TypeOf(info[1]).Elem().Name()
		MsgName2Id[msgName] = msgId
	}
}
