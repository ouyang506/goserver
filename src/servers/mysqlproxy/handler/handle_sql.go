package handler

import (
	"encoding/json"
	"fmt"
	"framework/log"
	"framework/proto/pb/ss"
	"framework/rpc"
	"mysqlproxy/mysqlmgr"
)

func (h *MessageHandler) HandleRpcReqExecuteSql(ctx rpc.Context, req *ss.ReqExecuteSql, resp *ss.RespExecuteSql) {
	reqJson, _ := json.Marshal(req)
	log.Debug("rcv ReqExecuteSql : %s", string(reqJson))
	defer func() {
		respJson, _ := json.Marshal(resp)
		log.Debug("response RespExecuteSql : %s", string(respJson))
	}()

	sql := req.GetSql()
	paramJson := req.GetParams()
	params := []any{}
	err := json.Unmarshal([]byte(paramJson), &params)
	if err != nil {
		log.Error("unmarshal params error: %v, param string: %v", err, paramJson)
		resp.ErrCode = new(int32)
		*resp.ErrCode = int32(ss.SsMysqlProxyError_invalid_sql_param)
		resp.ErrDesc = new(string)
		*resp.ErrDesc = fmt.Sprintf("unmarshal params error, %v", err)
		return
	}

	switch req.GetType() {
	case ss.DbOperType_oper_query:
		columns, result, err := mysqlmgr.Instance().QuerySql(sql, params...)
		if err != nil {
			log.Error("query sql error, %v", err)
			resp.ErrCode = new(int32)
			*resp.ErrCode = int32(ss.SsMysqlProxyError_execute_sql_failed)
			resp.ErrDesc = new(string)
			*resp.ErrDesc = err.Error()
			return
		}
		resp.QueryResult = &ss.MysqlQueryResult{}
		resp.QueryResult.Columns = columns
		rows := []*ss.MysqlResultRow{}
		for _, row := range result {
			rows = append(rows, &ss.MysqlResultRow{
				Values: row,
			})
		}
		resp.QueryResult.Rows = rows
	case ss.DbOperType_oper_execute:
		lastInsertId, rowsAffected, err := mysqlmgr.Instance().ExecuteSql(sql, params...)
		if err != nil {
			log.Error("execute sql error, %v", err)
			resp.ErrCode = new(int32)
			*resp.ErrCode = int32(ss.SsMysqlProxyError_execute_sql_failed)
			resp.ErrDesc = new(string)
			*resp.ErrDesc = err.Error()
			return
		}
		resp.LastInsertId = &lastInsertId
		resp.AffectedCount = &rowsAffected
	default:
		resp.ErrCode = new(int32)
		*resp.ErrCode = int32(ss.SsMysqlProxyError_invalid_oper_type)
		resp.ErrDesc = new(string)
		*resp.ErrDesc = "invalid operation type"
	}
}
