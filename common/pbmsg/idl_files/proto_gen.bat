set pwd=%cd%

cd ../../../tools

protoc --proto_path=../common/pbmsg/idl_files --go_out=../common  --go_opt=Mcommon.proto=pbmsg/ 	common.proto
protoc --proto_path=../common/pbmsg/idl_files --go_out=../common  --go_opt=Mgateway.proto=pbmsg/ 	gateway.proto

cd %pwd%

pause
exit
