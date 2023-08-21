// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: ss_mysql_proxy.proto

package ss

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

//消息ID定义[范围10001-10100]
type SsMysqlProxy int32

const (
	SsMysqlProxy_msg_id_req_execute_sql SsMysqlProxy = 10001
)

// Enum value maps for SsMysqlProxy.
var (
	SsMysqlProxy_name = map[int32]string{
		10001: "msg_id_req_execute_sql",
	}
	SsMysqlProxy_value = map[string]int32{
		"msg_id_req_execute_sql": 10001,
	}
)

func (x SsMysqlProxy) Enum() *SsMysqlProxy {
	p := new(SsMysqlProxy)
	*p = x
	return p
}

func (x SsMysqlProxy) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SsMysqlProxy) Descriptor() protoreflect.EnumDescriptor {
	return file_ss_mysql_proxy_proto_enumTypes[0].Descriptor()
}

func (SsMysqlProxy) Type() protoreflect.EnumType {
	return &file_ss_mysql_proxy_proto_enumTypes[0]
}

func (x SsMysqlProxy) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *SsMysqlProxy) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = SsMysqlProxy(num)
	return nil
}

// Deprecated: Use SsMysqlProxy.Descriptor instead.
func (SsMysqlProxy) EnumDescriptor() ([]byte, []int) {
	return file_ss_mysql_proxy_proto_rawDescGZIP(), []int{0}
}

//错误码
type SsMysqlProxyError int32

const (
	SsMysqlProxyError_invalid_oper_type  SsMysqlProxyError = 1 //非法的操作类型
	SsMysqlProxyError_invalid_sql_param  SsMysqlProxyError = 2 //sql查询参数错误
	SsMysqlProxyError_execute_sql_failed SsMysqlProxyError = 3 //执行sql返回错误
)

// Enum value maps for SsMysqlProxyError.
var (
	SsMysqlProxyError_name = map[int32]string{
		1: "invalid_oper_type",
		2: "invalid_sql_param",
		3: "execute_sql_failed",
	}
	SsMysqlProxyError_value = map[string]int32{
		"invalid_oper_type":  1,
		"invalid_sql_param":  2,
		"execute_sql_failed": 3,
	}
)

func (x SsMysqlProxyError) Enum() *SsMysqlProxyError {
	p := new(SsMysqlProxyError)
	*p = x
	return p
}

func (x SsMysqlProxyError) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SsMysqlProxyError) Descriptor() protoreflect.EnumDescriptor {
	return file_ss_mysql_proxy_proto_enumTypes[1].Descriptor()
}

func (SsMysqlProxyError) Type() protoreflect.EnumType {
	return &file_ss_mysql_proxy_proto_enumTypes[1]
}

func (x SsMysqlProxyError) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *SsMysqlProxyError) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = SsMysqlProxyError(num)
	return nil
}

// Deprecated: Use SsMysqlProxyError.Descriptor instead.
func (SsMysqlProxyError) EnumDescriptor() ([]byte, []int) {
	return file_ss_mysql_proxy_proto_rawDescGZIP(), []int{1}
}

// mysql操作类型
type DbOperType int32

const (
	DbOperType_oper_query   DbOperType = 1 // 查询
	DbOperType_oper_execute DbOperType = 2 // 新增、更新或者删除操作
)

// Enum value maps for DbOperType.
var (
	DbOperType_name = map[int32]string{
		1: "oper_query",
		2: "oper_execute",
	}
	DbOperType_value = map[string]int32{
		"oper_query":   1,
		"oper_execute": 2,
	}
)

func (x DbOperType) Enum() *DbOperType {
	p := new(DbOperType)
	*p = x
	return p
}

func (x DbOperType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DbOperType) Descriptor() protoreflect.EnumDescriptor {
	return file_ss_mysql_proxy_proto_enumTypes[2].Descriptor()
}

func (DbOperType) Type() protoreflect.EnumType {
	return &file_ss_mysql_proxy_proto_enumTypes[2]
}

func (x DbOperType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *DbOperType) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = DbOperType(num)
	return nil
}

// Deprecated: Use DbOperType.Descriptor instead.
func (DbOperType) EnumDescriptor() ([]byte, []int) {
	return file_ss_mysql_proxy_proto_rawDescGZIP(), []int{2}
}

//mysql数据库操作请求
type ReqExecuteSql struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type   *DbOperType `protobuf:"varint,1,opt,name=type,enum=ss.DbOperType" json:"type,omitempty"`
	Sql    *string     `protobuf:"bytes,2,opt,name=sql" json:"sql,omitempty"`
	Params *string     `protobuf:"bytes,3,opt,name=params" json:"params,omitempty"`
}

func (x *ReqExecuteSql) Reset() {
	*x = ReqExecuteSql{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ss_mysql_proxy_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReqExecuteSql) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReqExecuteSql) ProtoMessage() {}

func (x *ReqExecuteSql) ProtoReflect() protoreflect.Message {
	mi := &file_ss_mysql_proxy_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReqExecuteSql.ProtoReflect.Descriptor instead.
func (*ReqExecuteSql) Descriptor() ([]byte, []int) {
	return file_ss_mysql_proxy_proto_rawDescGZIP(), []int{0}
}

func (x *ReqExecuteSql) GetType() DbOperType {
	if x != nil && x.Type != nil {
		return *x.Type
	}
	return DbOperType_oper_query
}

func (x *ReqExecuteSql) GetSql() string {
	if x != nil && x.Sql != nil {
		return *x.Sql
	}
	return ""
}

func (x *ReqExecuteSql) GetParams() string {
	if x != nil && x.Params != nil {
		return *x.Params
	}
	return ""
}

//mysql数据库操作返回
type RespExecuteSql struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ErrCode       *int32            `protobuf:"varint,1,opt,name=err_code,json=errCode" json:"err_code,omitempty"`
	ErrDesc       *string           `protobuf:"bytes,2,opt,name=err_desc,json=errDesc" json:"err_desc,omitempty"`
	AffectedCount *int64            `protobuf:"varint,3,opt,name=affected_count,json=affectedCount" json:"affected_count,omitempty"`
	LastInsertId  *int64            `protobuf:"varint,4,opt,name=last_insert_id,json=lastInsertId" json:"last_insert_id,omitempty"`
	QueryResult   *MysqlQueryResult `protobuf:"bytes,5,opt,name=query_result,json=queryResult" json:"query_result,omitempty"`
}

func (x *RespExecuteSql) Reset() {
	*x = RespExecuteSql{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ss_mysql_proxy_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RespExecuteSql) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RespExecuteSql) ProtoMessage() {}

func (x *RespExecuteSql) ProtoReflect() protoreflect.Message {
	mi := &file_ss_mysql_proxy_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RespExecuteSql.ProtoReflect.Descriptor instead.
func (*RespExecuteSql) Descriptor() ([]byte, []int) {
	return file_ss_mysql_proxy_proto_rawDescGZIP(), []int{1}
}

func (x *RespExecuteSql) GetErrCode() int32 {
	if x != nil && x.ErrCode != nil {
		return *x.ErrCode
	}
	return 0
}

func (x *RespExecuteSql) GetErrDesc() string {
	if x != nil && x.ErrDesc != nil {
		return *x.ErrDesc
	}
	return ""
}

func (x *RespExecuteSql) GetAffectedCount() int64 {
	if x != nil && x.AffectedCount != nil {
		return *x.AffectedCount
	}
	return 0
}

func (x *RespExecuteSql) GetLastInsertId() int64 {
	if x != nil && x.LastInsertId != nil {
		return *x.LastInsertId
	}
	return 0
}

func (x *RespExecuteSql) GetQueryResult() *MysqlQueryResult {
	if x != nil {
		return x.QueryResult
	}
	return nil
}

type MysqlResultRow struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Values []string `protobuf:"bytes,1,rep,name=values" json:"values,omitempty"`
}

func (x *MysqlResultRow) Reset() {
	*x = MysqlResultRow{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ss_mysql_proxy_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MysqlResultRow) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MysqlResultRow) ProtoMessage() {}

func (x *MysqlResultRow) ProtoReflect() protoreflect.Message {
	mi := &file_ss_mysql_proxy_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MysqlResultRow.ProtoReflect.Descriptor instead.
func (*MysqlResultRow) Descriptor() ([]byte, []int) {
	return file_ss_mysql_proxy_proto_rawDescGZIP(), []int{2}
}

func (x *MysqlResultRow) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

type MysqlQueryResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Columns []string          `protobuf:"bytes,1,rep,name=columns" json:"columns,omitempty"`
	Rows    []*MysqlResultRow `protobuf:"bytes,2,rep,name=rows" json:"rows,omitempty"`
}

func (x *MysqlQueryResult) Reset() {
	*x = MysqlQueryResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ss_mysql_proxy_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MysqlQueryResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MysqlQueryResult) ProtoMessage() {}

func (x *MysqlQueryResult) ProtoReflect() protoreflect.Message {
	mi := &file_ss_mysql_proxy_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MysqlQueryResult.ProtoReflect.Descriptor instead.
func (*MysqlQueryResult) Descriptor() ([]byte, []int) {
	return file_ss_mysql_proxy_proto_rawDescGZIP(), []int{3}
}

func (x *MysqlQueryResult) GetColumns() []string {
	if x != nil {
		return x.Columns
	}
	return nil
}

func (x *MysqlQueryResult) GetRows() []*MysqlResultRow {
	if x != nil {
		return x.Rows
	}
	return nil
}

var File_ss_mysql_proxy_proto protoreflect.FileDescriptor

var file_ss_mysql_proxy_proto_rawDesc = []byte{
	0x0a, 0x14, 0x73, 0x73, 0x5f, 0x6d, 0x79, 0x73, 0x71, 0x6c, 0x5f, 0x70, 0x72, 0x6f, 0x78, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x73, 0x73, 0x22, 0x61, 0x0a, 0x0f, 0x72, 0x65,
	0x71, 0x5f, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x5f, 0x73, 0x71, 0x6c, 0x12, 0x24, 0x0a,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x10, 0x2e, 0x73, 0x73,
	0x2e, 0x64, 0x62, 0x5f, 0x6f, 0x70, 0x65, 0x72, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x73, 0x71, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x73, 0x71, 0x6c, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x22, 0xd0, 0x01,
	0x0a, 0x10, 0x72, 0x65, 0x73, 0x70, 0x5f, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x5f, 0x73,
	0x71, 0x6c, 0x12, 0x19, 0x0a, 0x08, 0x65, 0x72, 0x72, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x65, 0x72, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x19, 0x0a,
	0x08, 0x65, 0x72, 0x72, 0x5f, 0x64, 0x65, 0x73, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x65, 0x72, 0x72, 0x44, 0x65, 0x73, 0x63, 0x12, 0x25, 0x0a, 0x0e, 0x61, 0x66, 0x66, 0x65,
	0x63, 0x74, 0x65, 0x64, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x0d, 0x61, 0x66, 0x66, 0x65, 0x63, 0x74, 0x65, 0x64, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12,
	0x24, 0x0a, 0x0e, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x69, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x5f, 0x69,
	0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x6c, 0x61, 0x73, 0x74, 0x49, 0x6e, 0x73,
	0x65, 0x72, 0x74, 0x49, 0x64, 0x12, 0x39, 0x0a, 0x0c, 0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x72,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x73, 0x73,
	0x2e, 0x6d, 0x79, 0x73, 0x71, 0x6c, 0x5f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x72, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x52, 0x0b, 0x71, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x22, 0x2a, 0x0a, 0x10, 0x6d, 0x79, 0x73, 0x71, 0x6c, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x5f, 0x72, 0x6f, 0x77, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x22, 0x58, 0x0a, 0x12,
	0x6d, 0x79, 0x73, 0x71, 0x6c, 0x5f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x72, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x73, 0x12, 0x28, 0x0a, 0x04,
	0x72, 0x6f, 0x77, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x73, 0x73, 0x2e,
	0x6d, 0x79, 0x73, 0x71, 0x6c, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x5f, 0x72, 0x6f, 0x77,
	0x52, 0x04, 0x72, 0x6f, 0x77, 0x73, 0x2a, 0x2d, 0x0a, 0x0e, 0x73, 0x73, 0x5f, 0x6d, 0x79, 0x73,
	0x71, 0x6c, 0x5f, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x12, 0x1b, 0x0a, 0x16, 0x6d, 0x73, 0x67, 0x5f,
	0x69, 0x64, 0x5f, 0x72, 0x65, 0x71, 0x5f, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x5f, 0x73,
	0x71, 0x6c, 0x10, 0x91, 0x4e, 0x2a, 0x5c, 0x0a, 0x14, 0x73, 0x73, 0x5f, 0x6d, 0x79, 0x73, 0x71,
	0x6c, 0x5f, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x5f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x15, 0x0a,
	0x11, 0x69, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x5f, 0x6f, 0x70, 0x65, 0x72, 0x5f, 0x74, 0x79,
	0x70, 0x65, 0x10, 0x01, 0x12, 0x15, 0x0a, 0x11, 0x69, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x5f,
	0x73, 0x71, 0x6c, 0x5f, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x10, 0x02, 0x12, 0x16, 0x0a, 0x12, 0x65,
	0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x5f, 0x73, 0x71, 0x6c, 0x5f, 0x66, 0x61, 0x69, 0x6c, 0x65,
	0x64, 0x10, 0x03, 0x2a, 0x30, 0x0a, 0x0c, 0x64, 0x62, 0x5f, 0x6f, 0x70, 0x65, 0x72, 0x5f, 0x74,
	0x79, 0x70, 0x65, 0x12, 0x0e, 0x0a, 0x0a, 0x6f, 0x70, 0x65, 0x72, 0x5f, 0x71, 0x75, 0x65, 0x72,
	0x79, 0x10, 0x01, 0x12, 0x10, 0x0a, 0x0c, 0x6f, 0x70, 0x65, 0x72, 0x5f, 0x65, 0x78, 0x65, 0x63,
	0x75, 0x74, 0x65, 0x10, 0x02,
}

var (
	file_ss_mysql_proxy_proto_rawDescOnce sync.Once
	file_ss_mysql_proxy_proto_rawDescData = file_ss_mysql_proxy_proto_rawDesc
)

func file_ss_mysql_proxy_proto_rawDescGZIP() []byte {
	file_ss_mysql_proxy_proto_rawDescOnce.Do(func() {
		file_ss_mysql_proxy_proto_rawDescData = protoimpl.X.CompressGZIP(file_ss_mysql_proxy_proto_rawDescData)
	})
	return file_ss_mysql_proxy_proto_rawDescData
}

var file_ss_mysql_proxy_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_ss_mysql_proxy_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_ss_mysql_proxy_proto_goTypes = []interface{}{
	(SsMysqlProxy)(0),        // 0: ss.ss_mysql_proxy
	(SsMysqlProxyError)(0),   // 1: ss.ss_mysql_proxy_error
	(DbOperType)(0),          // 2: ss.db_oper_type
	(*ReqExecuteSql)(nil),    // 3: ss.req_execute_sql
	(*RespExecuteSql)(nil),   // 4: ss.resp_execute_sql
	(*MysqlResultRow)(nil),   // 5: ss.mysql_result_row
	(*MysqlQueryResult)(nil), // 6: ss.mysql_query_result
}
var file_ss_mysql_proxy_proto_depIdxs = []int32{
	2, // 0: ss.req_execute_sql.type:type_name -> ss.db_oper_type
	6, // 1: ss.resp_execute_sql.query_result:type_name -> ss.mysql_query_result
	5, // 2: ss.mysql_query_result.rows:type_name -> ss.mysql_result_row
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_ss_mysql_proxy_proto_init() }
func file_ss_mysql_proxy_proto_init() {
	if File_ss_mysql_proxy_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ss_mysql_proxy_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReqExecuteSql); i {
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
		file_ss_mysql_proxy_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RespExecuteSql); i {
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
		file_ss_mysql_proxy_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MysqlResultRow); i {
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
		file_ss_mysql_proxy_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MysqlQueryResult); i {
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
			RawDescriptor: file_ss_mysql_proxy_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ss_mysql_proxy_proto_goTypes,
		DependencyIndexes: file_ss_mysql_proxy_proto_depIdxs,
		EnumInfos:         file_ss_mysql_proxy_proto_enumTypes,
		MessageInfos:      file_ss_mysql_proxy_proto_msgTypes,
	}.Build()
	File_ss_mysql_proxy_proto = out.File
	file_ss_mysql_proxy_proto_rawDesc = nil
	file_ss_mysql_proxy_proto_goTypes = nil
	file_ss_mysql_proxy_proto_depIdxs = nil
}
