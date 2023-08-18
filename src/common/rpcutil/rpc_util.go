package rpcutil

import (
	"common"
	"framework/rpc"

	"google.golang.org/protobuf/proto"
)

// rpc to mysql proxy server
func CallMysqlProxy(req proto.Message, resp proto.Message, options ...rpc.Option) error {
	return rpc.Call(common.ServerTypeMysqlProxy, 0, req, resp, options...)
}
