package mysqlutil

import (
	"common/rpcutil"
	"encoding/json"
	"errors"
	"framework/log"
	"framework/proto/pb/ss"
	"framework/rpc"
	"time"
)

// select
func Query(sql string, params ...any) (result *QueryResult, err error) {
	req := &ss.ReqExecuteSql{}
	req.Type = new(ss.DbOperType)
	*req.Type = ss.DbOperType_oper_query
	req.Sql = &sql
	jsonBytes, _ := json.Marshal(params)
	req.Params = new(string)
	*req.Params = string(jsonBytes)
	resp := &ss.RespExecuteSql{}

	err = rpcutil.CallMysqlProxy(req, resp, rpc.WithTimeout(time.Second*5))
	if err != nil {
		log.Error("query sql error : %v, sql : %v, params: %v", err, sql, params)
		return
	}

	if resp.GetErrCode() != 0 {
		log.Error("query sql response error, sql : %v,  error code: %v, error desc : %v",
			sql, resp.GetErrCode(), resp.GetErrDesc())
		err = errors.New(resp.GetErrDesc())
		return
	}

	result = &QueryResult{}
	result.columns = resp.GetQueryResult().GetColumns()
	for _, row := range resp.GetQueryResult().GetRows() {
		resultRow := &ResultRow{columns: result.columns}
		for i := range row.Values {
			resultRow.fieldValues = append(resultRow.fieldValues, FieldValue(row.Values[i]))
		}
		result.rows = append(result.rows, resultRow)
	}
	return
}

// 查询一行数据，结尾自动添加limit 1
// 如果不需要添加后缀，使用Query()
func QueryOne(sql string, params ...any) (row *ResultRow, err error) {
	sql += " limit 1"
	result, err := Query(sql, params...)
	if err != nil {
		return nil, err
	}
	if result.RowCount() == 0 {
		return nil, nil
	}
	return result.Row(0), nil
}

// insert , update
func Execute(sql string, params ...any) (lastInsertId int64, affectCount int64, err error) {
	req := &ss.ReqExecuteSql{}
	req.Type = new(ss.DbOperType)
	*req.Type = ss.DbOperType_oper_execute
	req.Sql = &sql
	jsonBytes, _ := json.Marshal(params)
	req.Params = new(string)
	*req.Params = string(jsonBytes)
	resp := &ss.RespExecuteSql{}

	err = rpcutil.CallMysqlProxy(req, resp, rpc.WithTimeout(time.Second*5))
	if err != nil {
		log.Error("execute sql error : %v, sql : %v, params: %v", err, sql, params)
		return
	}

	if resp.GetErrCode() != 0 {
		log.Error("execute sql response error, sql : %v,  error code: %v, error desc : %v",
			sql, resp.GetErrCode(), resp.GetErrDesc())
		err = errors.New(resp.GetErrDesc())
		return
	}

	lastInsertId = resp.GetLastInsertId()
	affectCount = resp.GetAffectedCount()
	return
}
