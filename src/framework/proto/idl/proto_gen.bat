set pwd=%cd%

: protoc生成go胶水代码
cd ../../../../tools

set protoc_cs=protoc -I../src/framework/proto/idl/cs --go_out=../src/framework/proto/pb
%protoc_cs%  --go_opt=Mcs_common.proto=cs/ 	cs_common.proto
%protoc_cs%  --go_opt=Mcs_gate.proto=cs/ 	cs_gate.proto


set protoc_ss=protoc -I../src/framework/proto/idl/cs -I../src/framework/proto/idl/ss --go_out=../src/framework/proto/pb
%protoc_ss%  --go_opt=Mss_common.proto=ss/ 	ss_common.proto
%protoc_ss%  --go_opt=Mss_mysql_proxy.proto=ss/ 	ss_mysql_proxy.proto

cd %pwd%

: 生成rpc消息胶水代码rpc_msg_define.go
cd ../../../../tools/proto_parser
proto_parser
cd %pwd%

pause
exit
