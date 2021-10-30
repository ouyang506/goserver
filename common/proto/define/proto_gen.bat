set pwd=%cd%

cd ../../../tools

protoc --proto_path=../common/proto/define --go_out=../common  --go_opt=Mcommon.proto=proto/ 	common.proto
protoc --proto_path=../common/proto/define --go_out=../common  --go_opt=Mgateway.proto=proto/ 	gateway.proto

cd %pwd%

pause
exit
