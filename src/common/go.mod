//common主要是封装一些公用操作，供各个服务使用
//module依赖关系 common -> framework -> utility
module common

go 1.20

require (
	golang.org/x/exp v0.0.0-20230817173708-d852ddb80c63
	google.golang.org/protobuf v1.31.0
)
