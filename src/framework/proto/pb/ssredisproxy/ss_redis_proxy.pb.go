// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: ss_redis_proxy.proto

package ssredisproxy

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

//消息ID定义[范围10101-10200]
type MSGID int32

const (
	MSGID_msg_id_req_redis_cmd  MSGID = 10101
	MSGID_msg_id_req_redis_eval MSGID = 10102
)

// Enum value maps for MSGID.
var (
	MSGID_name = map[int32]string{
		10101: "msg_id_req_redis_cmd",
		10102: "msg_id_req_redis_eval",
	}
	MSGID_value = map[string]int32{
		"msg_id_req_redis_cmd":  10101,
		"msg_id_req_redis_eval": 10102,
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
	return file_ss_redis_proxy_proto_enumTypes[0].Descriptor()
}

func (MSGID) Type() protoreflect.EnumType {
	return &file_ss_redis_proxy_proto_enumTypes[0]
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
	return file_ss_redis_proxy_proto_rawDescGZIP(), []int{0}
}

//错误码
type Errors int32

const (
	Errors_redis_nil_error Errors = 1
	Errors_execute_failed  Errors = 2
)

// Enum value maps for Errors.
var (
	Errors_name = map[int32]string{
		1: "redis_nil_error",
		2: "execute_failed",
	}
	Errors_value = map[string]int32{
		"redis_nil_error": 1,
		"execute_failed":  2,
	}
)

func (x Errors) Enum() *Errors {
	p := new(Errors)
	*p = x
	return p
}

func (x Errors) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Errors) Descriptor() protoreflect.EnumDescriptor {
	return file_ss_redis_proxy_proto_enumTypes[1].Descriptor()
}

func (Errors) Type() protoreflect.EnumType {
	return &file_ss_redis_proxy_proto_enumTypes[1]
}

func (x Errors) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *Errors) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = Errors(num)
	return nil
}

// Deprecated: Use Errors.Descriptor instead.
func (Errors) EnumDescriptor() ([]byte, []int) {
	return file_ss_redis_proxy_proto_rawDescGZIP(), []int{1}
}

//redis操作请求
type ReqRedisCmd struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Args []string `protobuf:"bytes,1,rep,name=args" json:"args,omitempty"`
}

func (x *ReqRedisCmd) Reset() {
	*x = ReqRedisCmd{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ss_redis_proxy_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReqRedisCmd) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReqRedisCmd) ProtoMessage() {}

func (x *ReqRedisCmd) ProtoReflect() protoreflect.Message {
	mi := &file_ss_redis_proxy_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReqRedisCmd.ProtoReflect.Descriptor instead.
func (*ReqRedisCmd) Descriptor() ([]byte, []int) {
	return file_ss_redis_proxy_proto_rawDescGZIP(), []int{0}
}

func (x *ReqRedisCmd) GetArgs() []string {
	if x != nil {
		return x.Args
	}
	return nil
}

//redis操作返回
type RespRedisCmd struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ErrCode *int32  `protobuf:"varint,1,opt,name=err_code,json=errCode" json:"err_code,omitempty"`
	ErrDesc *string `protobuf:"bytes,2,opt,name=err_desc,json=errDesc" json:"err_desc,omitempty"`
	Result  *string `protobuf:"bytes,3,opt,name=result" json:"result,omitempty"` //json format
}

func (x *RespRedisCmd) Reset() {
	*x = RespRedisCmd{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ss_redis_proxy_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RespRedisCmd) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RespRedisCmd) ProtoMessage() {}

func (x *RespRedisCmd) ProtoReflect() protoreflect.Message {
	mi := &file_ss_redis_proxy_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RespRedisCmd.ProtoReflect.Descriptor instead.
func (*RespRedisCmd) Descriptor() ([]byte, []int) {
	return file_ss_redis_proxy_proto_rawDescGZIP(), []int{1}
}

func (x *RespRedisCmd) GetErrCode() int32 {
	if x != nil && x.ErrCode != nil {
		return *x.ErrCode
	}
	return 0
}

func (x *RespRedisCmd) GetErrDesc() string {
	if x != nil && x.ErrDesc != nil {
		return *x.ErrDesc
	}
	return ""
}

func (x *RespRedisCmd) GetResult() string {
	if x != nil && x.Result != nil {
		return *x.Result
	}
	return ""
}

type ReqRedisEval struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Script *string  `protobuf:"bytes,1,opt,name=script" json:"script,omitempty"`
	Keys   []string `protobuf:"bytes,2,rep,name=keys" json:"keys,omitempty"`
	Args   []string `protobuf:"bytes,3,rep,name=args" json:"args,omitempty"`
}

func (x *ReqRedisEval) Reset() {
	*x = ReqRedisEval{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ss_redis_proxy_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReqRedisEval) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReqRedisEval) ProtoMessage() {}

func (x *ReqRedisEval) ProtoReflect() protoreflect.Message {
	mi := &file_ss_redis_proxy_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReqRedisEval.ProtoReflect.Descriptor instead.
func (*ReqRedisEval) Descriptor() ([]byte, []int) {
	return file_ss_redis_proxy_proto_rawDescGZIP(), []int{2}
}

func (x *ReqRedisEval) GetScript() string {
	if x != nil && x.Script != nil {
		return *x.Script
	}
	return ""
}

func (x *ReqRedisEval) GetKeys() []string {
	if x != nil {
		return x.Keys
	}
	return nil
}

func (x *ReqRedisEval) GetArgs() []string {
	if x != nil {
		return x.Args
	}
	return nil
}

type RespRedisEval struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ErrCode *int32  `protobuf:"varint,1,opt,name=err_code,json=errCode" json:"err_code,omitempty"`
	ErrDesc *string `protobuf:"bytes,2,opt,name=err_desc,json=errDesc" json:"err_desc,omitempty"`
	Result  *string `protobuf:"bytes,3,opt,name=result" json:"result,omitempty"` //json format
}

func (x *RespRedisEval) Reset() {
	*x = RespRedisEval{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ss_redis_proxy_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RespRedisEval) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RespRedisEval) ProtoMessage() {}

func (x *RespRedisEval) ProtoReflect() protoreflect.Message {
	mi := &file_ss_redis_proxy_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RespRedisEval.ProtoReflect.Descriptor instead.
func (*RespRedisEval) Descriptor() ([]byte, []int) {
	return file_ss_redis_proxy_proto_rawDescGZIP(), []int{3}
}

func (x *RespRedisEval) GetErrCode() int32 {
	if x != nil && x.ErrCode != nil {
		return *x.ErrCode
	}
	return 0
}

func (x *RespRedisEval) GetErrDesc() string {
	if x != nil && x.ErrDesc != nil {
		return *x.ErrDesc
	}
	return ""
}

func (x *RespRedisEval) GetResult() string {
	if x != nil && x.Result != nil {
		return *x.Result
	}
	return ""
}

var File_ss_redis_proxy_proto protoreflect.FileDescriptor

var file_ss_redis_proxy_proto_rawDesc = []byte{
	0x0a, 0x14, 0x73, 0x73, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x73, 0x5f, 0x70, 0x72, 0x6f, 0x78, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x73, 0x73, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x73,
	0x5f, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x22, 0x23, 0x0a, 0x0d, 0x72, 0x65, 0x71, 0x5f, 0x72, 0x65,
	0x64, 0x69, 0x73, 0x5f, 0x63, 0x6d, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x72, 0x67, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x61, 0x72, 0x67, 0x73, 0x22, 0x5e, 0x0a, 0x0e, 0x72,
	0x65, 0x73, 0x70, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x73, 0x5f, 0x63, 0x6d, 0x64, 0x12, 0x19, 0x0a,
	0x08, 0x65, 0x72, 0x72, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x07, 0x65, 0x72, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x65, 0x72, 0x72, 0x5f,
	0x64, 0x65, 0x73, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x65, 0x72, 0x72, 0x44,
	0x65, 0x73, 0x63, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x50, 0x0a, 0x0e, 0x72,
	0x65, 0x71, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x73, 0x5f, 0x65, 0x76, 0x61, 0x6c, 0x12, 0x16, 0x0a,
	0x06, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x72, 0x67,
	0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x61, 0x72, 0x67, 0x73, 0x22, 0x5f, 0x0a,
	0x0f, 0x72, 0x65, 0x73, 0x70, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x73, 0x5f, 0x65, 0x76, 0x61, 0x6c,
	0x12, 0x19, 0x0a, 0x08, 0x65, 0x72, 0x72, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x07, 0x65, 0x72, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x65,
	0x72, 0x72, 0x5f, 0x64, 0x65, 0x73, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x65,
	0x72, 0x72, 0x44, 0x65, 0x73, 0x63, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x2a, 0x3e,
	0x0a, 0x05, 0x4d, 0x53, 0x47, 0x49, 0x44, 0x12, 0x19, 0x0a, 0x14, 0x6d, 0x73, 0x67, 0x5f, 0x69,
	0x64, 0x5f, 0x72, 0x65, 0x71, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x73, 0x5f, 0x63, 0x6d, 0x64, 0x10,
	0xf5, 0x4e, 0x12, 0x1a, 0x0a, 0x15, 0x6d, 0x73, 0x67, 0x5f, 0x69, 0x64, 0x5f, 0x72, 0x65, 0x71,
	0x5f, 0x72, 0x65, 0x64, 0x69, 0x73, 0x5f, 0x65, 0x76, 0x61, 0x6c, 0x10, 0xf6, 0x4e, 0x2a, 0x31,
	0x0a, 0x06, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x12, 0x13, 0x0a, 0x0f, 0x72, 0x65, 0x64, 0x69,
	0x73, 0x5f, 0x6e, 0x69, 0x6c, 0x5f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x10, 0x01, 0x12, 0x12, 0x0a,
	0x0e, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x5f, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x10,
	0x02,
}

var (
	file_ss_redis_proxy_proto_rawDescOnce sync.Once
	file_ss_redis_proxy_proto_rawDescData = file_ss_redis_proxy_proto_rawDesc
)

func file_ss_redis_proxy_proto_rawDescGZIP() []byte {
	file_ss_redis_proxy_proto_rawDescOnce.Do(func() {
		file_ss_redis_proxy_proto_rawDescData = protoimpl.X.CompressGZIP(file_ss_redis_proxy_proto_rawDescData)
	})
	return file_ss_redis_proxy_proto_rawDescData
}

var file_ss_redis_proxy_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_ss_redis_proxy_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_ss_redis_proxy_proto_goTypes = []interface{}{
	(MSGID)(0),            // 0: ss_redis_proxy.MSGID
	(Errors)(0),           // 1: ss_redis_proxy.Errors
	(*ReqRedisCmd)(nil),   // 2: ss_redis_proxy.req_redis_cmd
	(*RespRedisCmd)(nil),  // 3: ss_redis_proxy.resp_redis_cmd
	(*ReqRedisEval)(nil),  // 4: ss_redis_proxy.req_redis_eval
	(*RespRedisEval)(nil), // 5: ss_redis_proxy.resp_redis_eval
}
var file_ss_redis_proxy_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_ss_redis_proxy_proto_init() }
func file_ss_redis_proxy_proto_init() {
	if File_ss_redis_proxy_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ss_redis_proxy_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReqRedisCmd); i {
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
		file_ss_redis_proxy_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RespRedisCmd); i {
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
		file_ss_redis_proxy_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReqRedisEval); i {
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
		file_ss_redis_proxy_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RespRedisEval); i {
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
			RawDescriptor: file_ss_redis_proxy_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ss_redis_proxy_proto_goTypes,
		DependencyIndexes: file_ss_redis_proxy_proto_depIdxs,
		EnumInfos:         file_ss_redis_proxy_proto_enumTypes,
		MessageInfos:      file_ss_redis_proxy_proto_msgTypes,
	}.Build()
	File_ss_redis_proxy_proto = out.File
	file_ss_redis_proxy_proto_rawDesc = nil
	file_ss_redis_proxy_proto_goTypes = nil
	file_ss_redis_proxy_proto_depIdxs = nil
}