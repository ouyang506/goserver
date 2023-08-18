package handler

import (
	"encoding/json"
	"framework/log"
	"framework/proto/pb/ss"
	"mysqlproxy/dbmgr"
)

func (h *MessageHandler) HandleRpcReqExecuteSql(req *ss.ReqExecuteSql, resp *ss.RespExecuteSql) {
	reqJson, _ := json.Marshal(req)
	log.Debug("rcv ReqExecuteSql : %s", string(reqJson))
	defer func() {
		respJson, _ := json.Marshal(resp)
		log.Debug("response RespExecuteSql : %s", string(respJson))
	}()

	mysqlMgr := dbmgr.GetMysqlMgr()

	sql := req.GetSql()
	paramJson := req.GetParams()
	params := []any{}
	err := json.Unmarshal([]byte(paramJson), &params)
	if err != nil {
		log.Error("unmarshal params error: %v, param string: %v", err, paramJson)
		return
	}

	switch req.GetType() {
	case ss.DbOperType_oper_query:
		columns, result, err := mysqlMgr.QuerySql(sql, params...)
		if err != nil {
			log.Error("query sql error, %v", err)
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
		lastInsertId, rowsAffected, err := mysqlMgr.ExecuteSql(sql, params...)
		if err != nil {
			log.Error("execute sql error, %v", err)
			return
		}
		resp.LastInsertId = &lastInsertId
		resp.AffectedCount = &rowsAffected
	default:

	}
}

func (h *MessageHandler) HandleRpcNotifyExecuteSql(notify *ss.NotifyExecuteSql) {
	reqJson, _ := json.Marshal(notify)
	log.Debug("rcv NotifyExecuteSql : %s", string(reqJson))
}
