// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: cs_gate.proto

package cs

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

//消息ID定义[范围1000-1100]
type CsGate int32

const (
	//[1000,1010]定义系统协议
	CsGate_msg_id_req_heart_beat CsGate = 1001
	//[1011-1100]定义业务协议
	CsGate_msg_id_req_login_gate CsGate = 1011
	CsGate_msg_id_notify_tooltip CsGate = 1012
)

// Enum value maps for CsGate.
var (
	CsGate_name = map[int32]string{
		1001: "msg_id_req_heart_beat",
		1011: "msg_id_req_login_gate",
		1012: "msg_id_notify_tooltip",
	}
	CsGate_value = map[string]int32{
		"msg_id_req_heart_beat": 1001,
		"msg_id_req_login_gate": 1011,
		"msg_id_notify_tooltip": 1012,
	}
)

func (x CsGate) Enum() *CsGate {
	p := new(CsGate)
	*p = x
	return p
}

func (x CsGate) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CsGate) Descriptor() protoreflect.EnumDescriptor {
	return file_cs_gate_proto_enumTypes[0].Descriptor()
}

func (CsGate) Type() protoreflect.EnumType {
	return &file_cs_gate_proto_enumTypes[0]
}

func (x CsGate) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *CsGate) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = CsGate(num)
	return nil
}

// Deprecated: Use CsGate.Descriptor instead.
func (CsGate) EnumDescriptor() ([]byte, []int) {
	return file_cs_gate_proto_rawDescGZIP(), []int{0}
}

// 心跳请求协议
type ReqHeartBeat struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ReqHeartBeat) Reset() {
	*x = ReqHeartBeat{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cs_gate_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReqHeartBeat) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReqHeartBeat) ProtoMessage() {}

func (x *ReqHeartBeat) ProtoReflect() protoreflect.Message {
	mi := &file_cs_gate_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReqHeartBeat.ProtoReflect.Descriptor instead.
func (*ReqHeartBeat) Descriptor() ([]byte, []int) {
	return file_cs_gate_proto_rawDescGZIP(), []int{0}
}

// 心跳返回协议
type RespHeartBeat struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RespHeartBeat) Reset() {
	*x = RespHeartBeat{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cs_gate_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RespHeartBeat) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RespHeartBeat) ProtoMessage() {}

func (x *RespHeartBeat) ProtoReflect() protoreflect.Message {
	mi := &file_cs_gate_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RespHeartBeat.ProtoReflect.Descriptor instead.
func (*RespHeartBeat) Descriptor() ([]byte, []int) {
	return file_cs_gate_proto_rawDescGZIP(), []int{1}
}

//登录gateway请求
type ReqLoginGate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token *string `protobuf:"bytes,1,opt,name=token" json:"token,omitempty"`
}

func (x *ReqLoginGate) Reset() {
	*x = ReqLoginGate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cs_gate_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReqLoginGate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReqLoginGate) ProtoMessage() {}

func (x *ReqLoginGate) ProtoReflect() protoreflect.Message {
	mi := &file_cs_gate_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReqLoginGate.ProtoReflect.Descriptor instead.
func (*ReqLoginGate) Descriptor() ([]byte, []int) {
	return file_cs_gate_proto_rawDescGZIP(), []int{2}
}

func (x *ReqLoginGate) GetToken() string {
	if x != nil && x.Token != nil {
		return *x.Token
	}
	return ""
}

//登录gateway返回
type RespLoginGate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result *int32 `protobuf:"varint,1,opt,name=result" json:"result,omitempty"`
}

func (x *RespLoginGate) Reset() {
	*x = RespLoginGate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cs_gate_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RespLoginGate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RespLoginGate) ProtoMessage() {}

func (x *RespLoginGate) ProtoReflect() protoreflect.Message {
	mi := &file_cs_gate_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RespLoginGate.ProtoReflect.Descriptor instead.
func (*RespLoginGate) Descriptor() ([]byte, []int) {
	return file_cs_gate_proto_rawDescGZIP(), []int{3}
}

func (x *RespLoginGate) GetResult() int32 {
	if x != nil && x.Result != nil {
		return *x.Result
	}
	return 0
}

type NotifyTooltip struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Content *string `protobuf:"bytes,1,opt,name=content" json:"content,omitempty"`
}

func (x *NotifyTooltip) Reset() {
	*x = NotifyTooltip{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cs_gate_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotifyTooltip) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotifyTooltip) ProtoMessage() {}

func (x *NotifyTooltip) ProtoReflect() protoreflect.Message {
	mi := &file_cs_gate_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotifyTooltip.ProtoReflect.Descriptor instead.
func (*NotifyTooltip) Descriptor() ([]byte, []int) {
	return file_cs_gate_proto_rawDescGZIP(), []int{4}
}

func (x *NotifyTooltip) GetContent() string {
	if x != nil && x.Content != nil {
		return *x.Content
	}
	return ""
}

type TestEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TestEntry) Reset() {
	*x = TestEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cs_gate_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestEntry) ProtoMessage() {}

func (x *TestEntry) ProtoReflect() protoreflect.Message {
	mi := &file_cs_gate_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestEntry.ProtoReflect.Descriptor instead.
func (*TestEntry) Descriptor() ([]byte, []int) {
	return file_cs_gate_proto_rawDescGZIP(), []int{5}
}

var File_cs_gate_proto protoreflect.FileDescriptor

var file_cs_gate_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x63, 0x73, 0x5f, 0x67, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x02, 0x63, 0x73, 0x22, 0x10, 0x0a, 0x0e, 0x72, 0x65, 0x71, 0x5f, 0x68, 0x65, 0x61, 0x72, 0x74,
	0x5f, 0x62, 0x65, 0x61, 0x74, 0x22, 0x11, 0x0a, 0x0f, 0x72, 0x65, 0x73, 0x70, 0x5f, 0x68, 0x65,
	0x61, 0x72, 0x74, 0x5f, 0x62, 0x65, 0x61, 0x74, 0x22, 0x26, 0x0a, 0x0e, 0x72, 0x65, 0x71, 0x5f,
	0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x67, 0x61, 0x74, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f,
	0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e,
	0x22, 0x29, 0x0a, 0x0f, 0x72, 0x65, 0x73, 0x70, 0x5f, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x67,
	0x61, 0x74, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x2a, 0x0a, 0x0e, 0x6e,
	0x6f, 0x74, 0x69, 0x66, 0x79, 0x5f, 0x74, 0x6f, 0x6f, 0x6c, 0x74, 0x69, 0x70, 0x12, 0x18, 0x0a,
	0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x0c, 0x0a, 0x0a, 0x74, 0x65, 0x73, 0x74, 0x5f,
	0x65, 0x6e, 0x74, 0x72, 0x79, 0x2a, 0x5d, 0x0a, 0x07, 0x63, 0x73, 0x5f, 0x67, 0x61, 0x74, 0x65,
	0x12, 0x1a, 0x0a, 0x15, 0x6d, 0x73, 0x67, 0x5f, 0x69, 0x64, 0x5f, 0x72, 0x65, 0x71, 0x5f, 0x68,
	0x65, 0x61, 0x72, 0x74, 0x5f, 0x62, 0x65, 0x61, 0x74, 0x10, 0xe9, 0x07, 0x12, 0x1a, 0x0a, 0x15,
	0x6d, 0x73, 0x67, 0x5f, 0x69, 0x64, 0x5f, 0x72, 0x65, 0x71, 0x5f, 0x6c, 0x6f, 0x67, 0x69, 0x6e,
	0x5f, 0x67, 0x61, 0x74, 0x65, 0x10, 0xf3, 0x07, 0x12, 0x1a, 0x0a, 0x15, 0x6d, 0x73, 0x67, 0x5f,
	0x69, 0x64, 0x5f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x5f, 0x74, 0x6f, 0x6f, 0x6c, 0x74, 0x69,
	0x70, 0x10, 0xf4, 0x07,
}

var (
	file_cs_gate_proto_rawDescOnce sync.Once
	file_cs_gate_proto_rawDescData = file_cs_gate_proto_rawDesc
)

func file_cs_gate_proto_rawDescGZIP() []byte {
	file_cs_gate_proto_rawDescOnce.Do(func() {
		file_cs_gate_proto_rawDescData = protoimpl.X.CompressGZIP(file_cs_gate_proto_rawDescData)
	})
	return file_cs_gate_proto_rawDescData
}

var file_cs_gate_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_cs_gate_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_cs_gate_proto_goTypes = []interface{}{
	(CsGate)(0),           // 0: cs.cs_gate
	(*ReqHeartBeat)(nil),  // 1: cs.req_heart_beat
	(*RespHeartBeat)(nil), // 2: cs.resp_heart_beat
	(*ReqLoginGate)(nil),  // 3: cs.req_login_gate
	(*RespLoginGate)(nil), // 4: cs.resp_login_gate
	(*NotifyTooltip)(nil), // 5: cs.notify_tooltip
	(*TestEntry)(nil),     // 6: cs.test_entry
}
var file_cs_gate_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_cs_gate_proto_init() }
func file_cs_gate_proto_init() {
	if File_cs_gate_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cs_gate_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReqHeartBeat); i {
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
		file_cs_gate_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RespHeartBeat); i {
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
		file_cs_gate_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReqLoginGate); i {
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
		file_cs_gate_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RespLoginGate); i {
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
		file_cs_gate_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NotifyTooltip); i {
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
		file_cs_gate_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestEntry); i {
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
			RawDescriptor: file_cs_gate_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cs_gate_proto_goTypes,
		DependencyIndexes: file_cs_gate_proto_depIdxs,
		EnumInfos:         file_cs_gate_proto_enumTypes,
		MessageInfos:      file_cs_gate_proto_msgTypes,
	}.Build()
	File_cs_gate_proto = out.File
	file_cs_gate_proto_rawDesc = nil
	file_cs_gate_proto_goTypes = nil
	file_cs_gate_proto_depIdxs = nil
}
