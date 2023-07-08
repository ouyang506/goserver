set pwd=%cd%

: protoc生成go胶水代码
cd ../../../tools

set protoc_cs=protoc -I../common/proto/idl/cs --go_out=../common/proto/pb
%protoc_cs%  --go_opt=Mcs_head.proto=cs/ 	cs_head.proto
%protoc_cs%  --go_opt=Mcs_gate.proto=cs/ 	cs_gate.proto
%protoc_cs%  --go_opt=Mcs_rank.proto=cs/ 	cs_rank.proto


set protoc_ss=protoc -I../common/proto/idl/cs -I../common/proto/idl/ss --go_out=../common/proto/pb
%protoc_ss%  --go_opt=Mss_head.proto=ss/ 	ss_head.proto
%protoc_ss%  --go_opt=Mss_rank.proto=ss/ 	ss_rank.proto

cd %pwd%

:pause
exit
