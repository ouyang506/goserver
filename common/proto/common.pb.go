// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: common.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type MsgID int32

const (
	// common.proto
	MsgID_heart_beat_req  MsgID = 1
	MsgID_heart_beat_resp MsgID = 2
	// gateway.proto
	MsgID_login_gate_req  MsgID = 1000
	MsgID_login_gate_resp MsgID = 1001
)

// Enum value maps for MsgID.
var (
	MsgID_name = map[int32]string{
		1:    "heart_beat_req",
		2:    "heart_beat_resp",
		1000: "login_gate_req",
		1001: "login_gate_resp",
	}
	MsgID_value = map[string]int32{
		"heart_beat_req":  1,
		"heart_beat_resp": 2,
		"login_gate_req":  1000,
		"login_gate_resp": 1001,
	}
)

func (x MsgID) Enum() *MsgID {
	p := new(MsgID)
	*p = x
	return p
}

func (x MsgID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MsgID) Descriptor() protoreflect.EnumDescriptor {
	return file_common_proto_enumTypes[0].Descriptor()
}

func (MsgID) Type() protoreflect.EnumType {
	return &file_common_proto_enumTypes[0]
}

func (x MsgID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *MsgID) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = MsgID(num)
	return nil
}

// Deprecated: Use MsgID.Descriptor instead.
func (MsgID) EnumDescriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{0}
}

type MsgHead struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MsgHead) Reset() {
	*x = MsgHead{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgHead) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgHead) ProtoMessage() {}

func (x *MsgHead) ProtoReflect() protoreflect.Message {
	mi := &file_common_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MsgHead.ProtoReflect.Descriptor instead.
func (*MsgHead) Descriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{0}
}

type InnerMsgHead struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *InnerMsgHead) Reset() {
	*x = InnerMsgHead{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InnerMsgHead) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InnerMsgHead) ProtoMessage() {}

func (x *InnerMsgHead) ProtoReflect() protoreflect.Message {
	mi := &file_common_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InnerMsgHead.ProtoReflect.Descriptor instead.
func (*InnerMsgHead) Descriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{1}
}

// 心跳请求协议
type HeartBeatReqT struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *HeartBeatReqT) Reset() {
	*x = HeartBeatReqT{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HeartBeatReqT) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartBeatReqT) ProtoMessage() {}

func (x *HeartBeatReqT) ProtoReflect() protoreflect.Message {
	mi := &file_common_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeartBeatReqT.ProtoReflect.Descriptor instead.
func (*HeartBeatReqT) Descriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{2}
}

// 心跳返回协议
type HeartBeatRespT struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *HeartBeatRespT) Reset() {
	*x = HeartBeatRespT{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HeartBeatRespT) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartBeatRespT) ProtoMessage() {}

func (x *HeartBeatRespT) ProtoReflect() protoreflect.Message {
	mi := &file_common_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeartBeatRespT.ProtoReflect.Descriptor instead.
func (*HeartBeatRespT) Descriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{3}
}

var File_common_proto protoreflect.FileDescriptor

var file_common_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x22, 0x0a, 0x0a, 0x08, 0x6d, 0x73, 0x67, 0x5f, 0x68, 0x65,
	0x61, 0x64, 0x22, 0x10, 0x0a, 0x0e, 0x69, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x73, 0x67, 0x5f,
	0x68, 0x65, 0x61, 0x64, 0x22, 0x12, 0x0a, 0x10, 0x68, 0x65, 0x61, 0x72, 0x74, 0x5f, 0x62, 0x65,
	0x61, 0x74, 0x5f, 0x72, 0x65, 0x71, 0x5f, 0x74, 0x22, 0x13, 0x0a, 0x11, 0x68, 0x65, 0x61, 0x72,
	0x74, 0x5f, 0x62, 0x65, 0x61, 0x74, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x5f, 0x74, 0x2a, 0x5b, 0x0a,
	0x05, 0x4d, 0x73, 0x67, 0x49, 0x44, 0x12, 0x12, 0x0a, 0x0e, 0x68, 0x65, 0x61, 0x72, 0x74, 0x5f,
	0x62, 0x65, 0x61, 0x74, 0x5f, 0x72, 0x65, 0x71, 0x10, 0x01, 0x12, 0x13, 0x0a, 0x0f, 0x68, 0x65,
	0x61, 0x72, 0x74, 0x5f, 0x62, 0x65, 0x61, 0x74, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x10, 0x02, 0x12,
	0x13, 0x0a, 0x0e, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x67, 0x61, 0x74, 0x65, 0x5f, 0x72, 0x65,
	0x71, 0x10, 0xe8, 0x07, 0x12, 0x14, 0x0a, 0x0f, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x67, 0x61,
	0x74, 0x65, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x10, 0xe9, 0x07,
}

var (
	file_common_proto_rawDescOnce sync.Once
	file_common_proto_rawDescData = file_common_proto_rawDesc
)

func file_common_proto_rawDescGZIP() []byte {
	file_common_proto_rawDescOnce.Do(func() {
		file_common_proto_rawDescData = protoimpl.X.CompressGZIP(file_common_proto_rawDescData)
	})
	return file_common_proto_rawDescData
}

var file_common_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_common_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_common_proto_goTypes = []interface{}{
	(MsgID)(0),             // 0: common.MsgID
	(*MsgHead)(nil),        // 1: common.msg_head
	(*InnerMsgHead)(nil),   // 2: common.inner_msg_head
	(*HeartBeatReqT)(nil),  // 3: common.heart_beat_req_t
	(*HeartBeatRespT)(nil), // 4: common.heart_beat_resp_t
}
var file_common_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_common_proto_init() }
func file_common_proto_init() {
	if File_common_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_common_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgHead); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InnerMsgHead); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HeartBeatReqT); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HeartBeatRespT); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_common_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_proto_goTypes,
		DependencyIndexes: file_common_proto_depIdxs,
		EnumInfos:         file_common_proto_enumTypes,
		MessageInfos:      file_common_proto_msgTypes,
	}.Build()
	File_common_proto = out.File
	file_common_proto_rawDesc = nil
	file_common_proto_goTypes = nil
	file_common_proto_depIdxs = nil
}