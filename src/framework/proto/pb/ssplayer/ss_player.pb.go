// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: ss_player.proto

package ssplayer

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

//消息ID定义[范围10201-10300]
type MSGID int32

const (
	MSGID_msg_id_req_player_login  MSGID = 10201
	MSGID_msg_id_req_player_logout MSGID = 10202
)

// Enum value maps for MSGID.
var (
	MSGID_name = map[int32]string{
		10201: "msg_id_req_player_login",
		10202: "msg_id_req_player_logout",
	}
	MSGID_value = map[string]int32{
		"msg_id_req_player_login":  10201,
		"msg_id_req_player_logout": 10202,
	}
)

func (x MSGID) Enum() *MSGID {
	p := new(MSGID)
	*p = x
	return p
}

func (x MSGID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MSGID) Descriptor() protoreflect.EnumDescriptor {
	return file_ss_player_proto_enumTypes[0].Descriptor()
}

func (MSGID) Type() protoreflect.EnumType {
	return &file_ss_player_proto_enumTypes[0]
}

func (x MSGID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *MSGID) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = MSGID(num)
	return nil
}

// Deprecated: Use MSGID.Descriptor instead.
func (MSGID) EnumDescriptor() ([]byte, []int) {
	return file_ss_player_proto_rawDescGZIP(), []int{0}
}

type ReqPlayerLogin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PlayerId *int64 `protobuf:"varint,1,opt,name=player_id,json=playerId" json:"player_id,omitempty"`
}

func (x *ReqPlayerLogin) Reset() {
	*x = ReqPlayerLogin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ss_player_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReqPlayerLogin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReqPlayerLogin) ProtoMessage() {}

func (x *ReqPlayerLogin) ProtoReflect() protoreflect.Message {
	mi := &file_ss_player_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReqPlayerLogin.ProtoReflect.Descriptor instead.
func (*ReqPlayerLogin) Descriptor() ([]byte, []int) {
	return file_ss_player_proto_rawDescGZIP(), []int{0}
}

func (x *ReqPlayerLogin) GetPlayerId() int64 {
	if x != nil && x.PlayerId != nil {
		return *x.PlayerId
	}
	return 0
}

type RespPlayerLogin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ErrCode *int32  `protobuf:"varint,1,opt,name=err_code,json=errCode" json:"err_code,omitempty"`
	ErrDesc *string `protobuf:"bytes,2,opt,name=err_desc,json=errDesc" json:"err_desc,omitempty"`
}

func (x *RespPlayerLogin) Reset() {
	*x = RespPlayerLogin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ss_player_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RespPlayerLogin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RespPlayerLogin) ProtoMessage() {}

func (x *RespPlayerLogin) ProtoReflect() protoreflect.Message {
	mi := &file_ss_player_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RespPlayerLogin.ProtoReflect.Descriptor instead.
func (*RespPlayerLogin) Descriptor() ([]byte, []int) {
	return file_ss_player_proto_rawDescGZIP(), []int{1}
}

func (x *RespPlayerLogin) GetErrCode() int32 {
	if x != nil && x.ErrCode != nil {
		return *x.ErrCode
	}
	return 0
}

func (x *RespPlayerLogin) GetErrDesc() string {
	if x != nil && x.ErrDesc != nil {
		return *x.ErrDesc
	}
	return ""
}

type ReqPlayerLogout struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PlayerId *int64 `protobuf:"varint,1,opt,name=player_id,json=playerId" json:"player_id,omitempty"`
}

func (x *ReqPlayerLogout) Reset() {
	*x = ReqPlayerLogout{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ss_player_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReqPlayerLogout) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReqPlayerLogout) ProtoMessage() {}

func (x *ReqPlayerLogout) ProtoReflect() protoreflect.Message {
	mi := &file_ss_player_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReqPlayerLogout.ProtoReflect.Descriptor instead.
func (*ReqPlayerLogout) Descriptor() ([]byte, []int) {
	return file_ss_player_proto_rawDescGZIP(), []int{2}
}

func (x *ReqPlayerLogout) GetPlayerId() int64 {
	if x != nil && x.PlayerId != nil {
		return *x.PlayerId
	}
	return 0
}

type RespPlayerLogout struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ErrCode *int32  `protobuf:"varint,1,opt,name=err_code,json=errCode" json:"err_code,omitempty"`
	ErrDesc *string `protobuf:"bytes,2,opt,name=err_desc,json=errDesc" json:"err_desc,omitempty"`
}

func (x *RespPlayerLogout) Reset() {
	*x = RespPlayerLogout{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ss_player_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RespPlayerLogout) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RespPlayerLogout) ProtoMessage() {}

func (x *RespPlayerLogout) ProtoReflect() protoreflect.Message {
	mi := &file_ss_player_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RespPlayerLogout.ProtoReflect.Descriptor instead.
func (*RespPlayerLogout) Descriptor() ([]byte, []int) {
	return file_ss_player_proto_rawDescGZIP(), []int{3}
}

func (x *RespPlayerLogout) GetErrCode() int32 {
	if x != nil && x.ErrCode != nil {
		return *x.ErrCode
	}
	return 0
}

func (x *RespPlayerLogout) GetErrDesc() string {
	if x != nil && x.ErrDesc != nil {
		return *x.ErrDesc
	}
	return ""
}

var File_ss_player_proto protoreflect.FileDescriptor

var file_ss_player_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x73, 0x73, 0x5f, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x09, 0x73, 0x73, 0x5f, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x22, 0x2f, 0x0a, 0x10,
	0x72, 0x65, 0x71, 0x5f, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x5f, 0x6c, 0x6f, 0x67, 0x69, 0x6e,
	0x12, 0x1b, 0x0a, 0x09, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x08, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x49, 0x64, 0x22, 0x49, 0x0a,
	0x11, 0x72, 0x65, 0x73, 0x70, 0x5f, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x5f, 0x6c, 0x6f, 0x67,
	0x69, 0x6e, 0x12, 0x19, 0x0a, 0x08, 0x65, 0x72, 0x72, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x65, 0x72, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x19, 0x0a,
	0x08, 0x65, 0x72, 0x72, 0x5f, 0x64, 0x65, 0x73, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x65, 0x72, 0x72, 0x44, 0x65, 0x73, 0x63, 0x22, 0x30, 0x0a, 0x11, 0x72, 0x65, 0x71, 0x5f,
	0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x5f, 0x6c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x12, 0x1b, 0x0a,
	0x09, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x08, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x49, 0x64, 0x22, 0x4a, 0x0a, 0x12, 0x72, 0x65,
	0x73, 0x70, 0x5f, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x5f, 0x6c, 0x6f, 0x67, 0x6f, 0x75, 0x74,
	0x12, 0x19, 0x0a, 0x08, 0x65, 0x72, 0x72, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x07, 0x65, 0x72, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x65,
	0x72, 0x72, 0x5f, 0x64, 0x65, 0x73, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x65,
	0x72, 0x72, 0x44, 0x65, 0x73, 0x63, 0x2a, 0x44, 0x0a, 0x05, 0x4d, 0x53, 0x47, 0x49, 0x44, 0x12,
	0x1c, 0x0a, 0x17, 0x6d, 0x73, 0x67, 0x5f, 0x69, 0x64, 0x5f, 0x72, 0x65, 0x71, 0x5f, 0x70, 0x6c,
	0x61, 0x79, 0x65, 0x72, 0x5f, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x10, 0xd9, 0x4f, 0x12, 0x1d, 0x0a,
	0x18, 0x6d, 0x73, 0x67, 0x5f, 0x69, 0x64, 0x5f, 0x72, 0x65, 0x71, 0x5f, 0x70, 0x6c, 0x61, 0x79,
	0x65, 0x72, 0x5f, 0x6c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x10, 0xda, 0x4f,
}

var (
	file_ss_player_proto_rawDescOnce sync.Once
	file_ss_player_proto_rawDescData = file_ss_player_proto_rawDesc
)

func file_ss_player_proto_rawDescGZIP() []byte {
	file_ss_player_proto_rawDescOnce.Do(func() {
		file_ss_player_proto_rawDescData = protoimpl.X.CompressGZIP(file_ss_player_proto_rawDescData)
	})
	return file_ss_player_proto_rawDescData
}

var file_ss_player_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_ss_player_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_ss_player_proto_goTypes = []interface{}{
	(MSGID)(0),               // 0: ss_player.MSGID
	(*ReqPlayerLogin)(nil),   // 1: ss_player.req_player_login
	(*RespPlayerLogin)(nil),  // 2: ss_player.resp_player_login
	(*ReqPlayerLogout)(nil),  // 3: ss_player.req_player_logout
	(*RespPlayerLogout)(nil), // 4: ss_player.resp_player_logout
}
var file_ss_player_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_ss_player_proto_init() }
func file_ss_player_proto_init() {
	if File_ss_player_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ss_player_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReqPlayerLogin); i {
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
		file_ss_player_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RespPlayerLogin); i {
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
		file_ss_player_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReqPlayerLogout); i {
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
		file_ss_player_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RespPlayerLogout); i {
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
			RawDescriptor: file_ss_player_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ss_player_proto_goTypes,
		DependencyIndexes: file_ss_player_proto_depIdxs,
		EnumInfos:         file_ss_player_proto_enumTypes,
		MessageInfos:      file_ss_player_proto_msgTypes,
	}.Build()
	File_ss_player_proto = out.File
	file_ss_player_proto_rawDesc = nil
	file_ss_player_proto_goTypes = nil
	file_ss_player_proto_depIdxs = nil
}
