package mysqlutil

import (
	"common/rpcutil"
	"encoding/json"
	"framework/log"
	"framework/proto/pb/ss"
	"strconv"
)

type MysqlResultRow []MysqlFieldValue
type MysqlFieldValue string
type MysqlQueryResult struct {
	Columns []string
	Rows    []MysqlResultRow
}

func (fv *MysqlFieldValue) String() string {
	return string(*fv)
}

func (fv *MysqlFieldValue) Int() int {
	return int(fv.Int64())
}

func (fv *MysqlFieldValue) Uint() uint {
	return uint(fv.Uint64())
}

func (fv *MysqlFieldValue) Int32() int32 {
	return int32(fv.Int64())
}

func (fv *MysqlFieldValue) Uint32() uint32 {
	return uint32(fv.Uint64())
}

func (fv *MysqlFieldValue) Int64() int64 {
	i, _ := strconv.ParseInt(string(*fv), 10, 64)
	return i
}

func (fv *MysqlFieldValue) Uint64() uint64 {
	i, _ := strconv.ParseUint(string(*fv), 10, 64)
	return i
}

func QuerySQL(sql string, params ...any) (result *MysqlQueryResult, err error) {
	req := &ss.ReqExecuteSql{}
	req.Type = new(ss.DbOperType)
	*req.Type = ss.DbOperType_oper_query
	req.Sql = &sql
	jsonBytes, _ := json.Marshal(params)
	req.Params = new(string)
	*req.Params = string(jsonBytes)
	resp := &ss.RespExecuteSql{}

	err = rpcutil.CallMysqlProxy(req, resp)
	if err != nil {
		log.Error("query sql error : %v, sql : %v, params: %v", err, sql, params)
		return
	}

	result = &MysqlQueryResult{}
	result.Columns = resp.GetQueryResult().GetColumns()
	for _, row := range resp.GetQueryResult().GetRows() {
		r := make([]MysqlFieldValue, len(row.Values))
		for i := range row.Values {
			r[i] = MysqlFieldValue(row.Values[i])
		}
		result.Rows = append(result.Rows, r)
	}
	return
}
