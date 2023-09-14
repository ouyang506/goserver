set pwd=%cd%

: protoc生成go胶水代码
cd ../../../../tools

protoc -I../src/framework/proto/idl/ --go_out=../src/framework/proto --go_opt=Merror_code.proto=pb/ 	error_code.proto

set protoc_cs=protoc -I../src/framework/proto/idl/cs --go_out=../src/framework/proto/pb
%protoc_cs%  --go_opt=Mcs_common.proto=cscommon/ 	cs_common.proto
%protoc_cs%  --go_opt=Mcs_gate.proto=csgate/ 	cs_gate.proto
%protoc_cs%  --go_opt=Mcs_player.proto=csplayer/ 	cs_player.proto


set protoc_ss=protoc -I../src/framework/proto/idl/cs -I../src/framework/proto/idl/ss --go_out=../src/framework/proto/pb
%protoc_ss%  --go_opt=Mss_common.proto=sscommon/ 	ss_common.proto
%protoc_ss%  --go_opt=Mss_mysql_proxy.proto=ssmysqlproxy/ 	ss_mysql_proxy.proto
%protoc_ss%  --go_opt=Mss_redis_proxy.proto=ssredisproxy/ 	ss_redis_proxy.proto
%protoc_ss%  --go_opt=Mss_player.proto=ssplayer/ 	ss_player.proto

cd %pwd%

: 生成rpc消息胶水代码rpc_msg_define.go
cd ../../../../tools/proto_parser
proto_parser
cd %pwd%

rem pause
exit
