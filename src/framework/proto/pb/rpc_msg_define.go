// Code generated by tool. DO NOT EDIT.
package pb

import (
	"framework/proto/pb/csgate"
	"framework/proto/pb/csplayer"
	"framework/proto/pb/ssmysqlproxy"
	"framework/proto/pb/ssplayer"
	"framework/proto/pb/ssredisproxy"
)

var CSRpcMsg [][]any = [][]any{
	//cs_gate
	{int(csgate.MSGID_msg_id_req_heart_beat), &csgate.ReqHeartBeat{}, &csgate.RespHeartBeat{}},
	{int(csgate.MSGID_msg_id_req_login_gate), &csgate.ReqLoginGate{}, &csgate.RespLoginGate{}},
	//cs_player
	{int(csplayer.MSGID_msg_id_req_query_player), &csplayer.ReqQueryPlayer{}, &csplayer.RespQueryPlayer{}},
}

var SSRpcMsg [][]any = [][]any{
	//ss_mysql_proxy
	{int(ssmysqlproxy.MSGID_msg_id_req_execute_sql), &ssmysqlproxy.ReqExecuteSql{}, &ssmysqlproxy.RespExecuteSql{}},
	//ss_redis_proxy
	{int(ssredisproxy.MSGID_msg_id_req_redis_cmd), &ssredisproxy.ReqRedisCmd{}, &ssredisproxy.RespRedisCmd{}},
	{int(ssredisproxy.MSGID_msg_id_req_redis_eval), &ssredisproxy.ReqRedisEval{}, &ssredisproxy.RespRedisEval{}},
	//ss_player
	{int(ssplayer.MSGID_msg_id_req_player_login), &ssplayer.ReqPlayerLogin{}, &ssplayer.RespPlayerLogin{}},
	{int(ssplayer.MSGID_msg_id_req_player_logout), &ssplayer.ReqPlayerLogout{}, &ssplayer.RespPlayerLogout{}},
}
