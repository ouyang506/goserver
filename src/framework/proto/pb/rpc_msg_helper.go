package pb

import (
	"reflect"

	"google.golang.org/protobuf/proto"
)

var (
	MsgId2Type map[int][]proto.Message = make(map[int][]proto.Message)
	MsgId2Name map[int]string          = make(map[int]string)
	MsgName2Id map[string]int          = make(map[string]int)
)

func init() {
	fillMsgMap()
}

func fillMsgMap() {
	for _, info := range CSRpcMsg {
		msgId := info[0].(int)
		arr := []proto.Message{}
		arr = append(arr, proto.Clone(info[1].(proto.Message)))
		if len(info) >= 3 && info[2] != nil {
			arr = append(arr, proto.Clone(info[2].(proto.Message)))
		}
		MsgId2Type[msgId] = arr

		msgName := reflect.TypeOf(info[1]).Elem().Name()
		MsgName2Id[msgName] = msgId
		MsgId2Name[msgId] = msgName
	}

	for _, info := range SSRpcMsg {
		msgId := info[0].(int)
		arr := []proto.Message{}
		arr = append(arr, proto.Clone(info[1].(proto.Message)))
		if len(info) >= 3 && info[2] != nil {
			arr = append(arr, proto.Clone(info[2].(proto.Message)))
		}
		MsgId2Type[msgId] = arr

		msgName := reflect.TypeOf(info[1]).Elem().Name()
		MsgName2Id[msgName] = msgId
		MsgId2Name[msgId] = msgName
	}
}

// TODO: need a message pool here
func GetProtoMsgById(msgId int) (req proto.Message, resp proto.Message) {
	msgArr, ok := MsgId2Type[msgId]
	if !ok {
		return
	}
	if len(msgArr) >= 1 {
		req = proto.Clone(msgArr[0])
	}
	if len(msgArr) >= 2 {
		resp = proto.Clone(msgArr[1])
	}
	return
}

func GetMsgIdByName(msgName string) int {
	msgId, ok := MsgName2Id[msgName]
	if !ok {
		return 0
	}
	return msgId
}

func GetMsgNameById(msgId int) string {
	name, ok := MsgId2Name[msgId]
	if !ok {
		return ""
	}
	return name
}
