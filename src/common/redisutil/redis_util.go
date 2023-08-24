package redisutil

import (
	"common/rpcutil"
	"encoding/json"
	"errors"
	"fmt"
	"framework/log"
	"framework/proto/pb/ss"
	"framework/rpc"
	"strconv"
	"time"
)

var (
	ErrRpc           = errors.New("rpc error")
	ErrRedisNil      = errors.New("redis nil")
	ErrRedisExecute  = errors.New("redis execute failed")
	ErrParseResponse = errors.New("unmarshal response failed")
)

func doCallRedis(cmdArgs []string, cmdResult any) error {
	req := &ss.ReqRedisCmd{}
	req.Args = cmdArgs

	resp := &ss.RespRedisCmd{}
	err := rpcutil.CallRedisProxy(req, resp, rpc.WithTimeout(time.Second*5))
	if err != nil {
		log.Error("rpc call redis proxy failed, %v", err)
		return ErrRpc
	}

	if resp.GetErrCode() != 0 {
		if resp.GetErrCode() == int32(ss.SsRedisProxyError_redis_nil_error) {
			return ErrRedisNil
		}
		log.Error("do redis cmd failed, %v", resp.GetErrDesc())
		return ErrRedisExecute
	}

	if cmdResult != nil {
		err = json.Unmarshal([]byte(resp.GetResult()), cmdResult)
		if err != nil {
			log.Error("unmarhsal redis response data failed, %v", err)
			return ErrParseResponse
		}
	}
	return nil
}

func Set(k, v string) error {
	cmdArgs := []string{"set", k, v}
	return doCallRedis(cmdArgs, nil)
}

func SetEx(k, v string, ttl int) error {
	cmdArgs := []string{"set", k, v, "ex", strconv.Itoa(ttl)}
	return doCallRedis(cmdArgs, nil)
}

func SetNx(k, v string, ttl int) error {
	var cmdArgs []string
	if ttl <= 0 {
		cmdArgs = []string{"set", k, v, "nx"}
	} else {
		cmdArgs = []string{"set", k, v, "ex", strconv.Itoa(ttl), "nx"}
	}
	return doCallRedis(cmdArgs, nil)
}

func SetXx(k, v string, ttl int) error {
	var cmdArgs []string
	if ttl <= 0 {
		cmdArgs = []string{"set", k, v, "xx"}
	} else {
		cmdArgs = []string{"set", k, v, "ex", strconv.Itoa(ttl), "xx"}
	}
	return doCallRedis(cmdArgs, nil)
}

func Get(k string) (string, error) {
	cmdArgs := []string{"get", k}
	result := new(string)
	err := doCallRedis(cmdArgs, result)
	if err != nil {
		if errors.Is(err, ErrRedisNil) {
			return "", nil
		}
		return "", err
	}
	return *result, nil
}

func Del(keys ...string) error {
	cmdArgs := []string{"del"}
	cmdArgs = append(cmdArgs, keys...)
	return doCallRedis(cmdArgs, nil)
}

func HSet(k, f, v string) error {
	cmdArgs := []string{"hset", k, f, v}
	return doCallRedis(cmdArgs, nil)
}

func HSetNx(k, f, v string) (bool, error) {
	cmdArgs := []string{"hsetnx", k, f, v}
	result := new(int)
	err := doCallRedis(cmdArgs, result)
	if err != nil {
		return false, err
	}
	return *result > 0, nil
}

func HMSet(k string, fv map[string]string) error {
	cmdArgs := []string{"hmset"}
	for f, v := range fv {
		cmdArgs = append(cmdArgs, f)
		cmdArgs = append(cmdArgs, v)
	}
	return doCallRedis(cmdArgs, nil)
}

func HGet(k, f string) (string, error) {
	cmdArgs := []string{"hget", k, f}
	result := new(string)
	err := doCallRedis(cmdArgs, result)
	if err != nil {
		if errors.Is(err, ErrRedisNil) {
			return "", nil
		}
		return "", err
	}
	return *result, nil
}

func HMGet(k string, f ...string) ([]string, error) {
	cmdArgs := []string{"hmget", k}
	cmdArgs = append(cmdArgs, f...)
	result := make([]any, len(f))
	err := doCallRedis(cmdArgs, &result)
	if err != nil {
		return nil, err
	}
	strResult := make([]string, len(result))
	for i := range result {
		switch v := result[i].(type) {
		case nil:
			strResult[i] = ""
		case string:
			strResult[i] = v
		default:
			strResult[i] = fmt.Sprintf("%v", result[i])
		}
	}
	return strResult, nil
}

func HGetAll(k string) (map[string]string, error) {
	cmdArgs := []string{"hgetall", k}
	result := []any{}
	err := doCallRedis(cmdArgs, &result)
	if err != nil {
		return nil, err
	}
	strResult := make([]string, len(result))
	for i := range result {
		switch v := result[i].(type) {
		case nil:
			strResult[i] = ""
		case string:
			strResult[i] = v
		default:
			strResult[i] = fmt.Sprintf("%v", result[i])
		}
	}
	mapResult := make(map[string]string)
	for i := 0; i < len(strResult)/2; i++ {
		mapResult[strResult[i*2]] = strResult[i*2+1]
	}
	return mapResult, nil
}

func HDel(k string, f ...string) error {
	cmdArgs := []string{"hdel", k}
	cmdArgs = append(cmdArgs, f...)
	return doCallRedis(cmdArgs, nil)
}
