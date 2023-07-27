set pwd=%cd%

: protoc生成go胶水代码
cd ../../../../tools

set protoc_cs=protoc -I../src/framework/proto/idl/cs --go_out=../src/framework/proto/pb
%protoc_cs%  --go_opt=Mcs_head.proto=cs/ 	cs_head.proto
%protoc_cs%  --go_opt=Mcs_gate.proto=cs/ 	cs_gate.proto
%protoc_cs%  --go_opt=Mcs_rank.proto=cs/ 	cs_rank.proto


set protoc_ss=protoc -I../src/framework/proto/idl/cs -I../src/framework/proto/idl/ss --go_out=../src/framework/proto/pb
%protoc_ss%  --go_opt=Mss_head.proto=ss/ 	ss_head.proto
%protoc_ss%  --go_opt=Mss_rank.proto=ss/ 	ss_rank.proto

cd %pwd%

: 生成rpc消息胶水代码rpc_msg_define.go
cd ../../../../tools/proto_parser
proto_parser
cd %pwd%

:pause
exit
