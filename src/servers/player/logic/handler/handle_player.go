package handler

import (
	"encoding/json"
	"framework/log"
	"framework/proto/pb"
	"framework/proto/pb/csplayer"
	"framework/proto/pb/ssplayer"
	"framework/rpc"
	"player/logic/playermgr"
	"time"
)

func (h *MessageHandler) HandleRpcReqPlayerLogin(ctx rpc.Context,
	req *ssplayer.ReqPlayerLogin, resp *ssplayer.RespPlayerLogin) {

	defer func() {
		respJson, _ := json.Marshal(resp)
		reqJson, _ := json.Marshal(req)
		log.Debug("rcv ReqPlayerLogin: %s, response RespPlayerLogin: %s",
			string(reqJson), string(respJson))
	}()

	playerLoginReq := &playermgr.PlayerLoginReq{
		PlayerId: req.GetPlayerId(),
	}
	future := h.Root().Request(playermgr.ActorId, playerLoginReq)
	_, err := future.WaitTimeout(time.Second * 5)
	if err != nil {
		log.Error("player login error: %v, playerId=%v", err, req.GetPlayerId())
		return
	}
}

func (h *MessageHandler) HandleRpcReqPlayerLogout(ctx rpc.Context,
	req *ssplayer.ReqPlayerLogout, resp *ssplayer.RespPlayerLogout) {

	defer func() {
		respJson, _ := json.Marshal(resp)
		reqJson, _ := json.Marshal(req)
		log.Debug("rcv ReqPlayerLogout: %s, response RespPlayerLogout: %s",
			string(reqJson), string(respJson))
	}()

	playerLogoutReq := &playermgr.PlayerLogoutReq{
		PlayerId: req.GetPlayerId(),
	}
	future := h.Root().Request(playermgr.ActorId, playerLogoutReq)
	_, err := future.WaitTimeout(time.Second * 5)
	if err != nil {
		log.Error("player logout error: %v, playerId=%v", err, req.GetPlayerId())
		return
	}
}

func (h *MessageHandler) HandleRpcReqQueryPlayer(ctx rpc.Context,
	req *csplayer.ReqQueryPlayer, resp *csplayer.RespQueryPlayer) {

	defer func() {
		respJson, _ := json.Marshal(resp)
		reqJson, _ := json.Marshal(req)
		log.Debug("rcv ReqQueryPlayer: %s, response RespQueryPlayer: %s, playerId: %v",
			string(reqJson), string(respJson), ctx.GetGuid())
	}()

	playerId := ctx.GetGuid()
	if playerId <= 0 {
		log.Error("get player guid error")
		resp.ErrCode = new(int32)
		*resp.ErrCode = int32(pb.ERROR_CODE_UNKOWN)
		resp.ErrDesc = new(string)
		*resp.ErrDesc = "get player guid failed"
		return
	}

	fetchReq := &playermgr.FetchPlayerReq{
		PlayerId: playerId,
	}
	future := h.Root().Request(playermgr.ActorId, fetchReq)
	_, err := future.WaitTimeout(time.Second * 5)
	if err != nil {
		log.Error("get player error, playerId=%v, err=%v", playerId, err)
		resp.ErrCode = new(int32)
		*resp.ErrCode = int32(pb.ERROR_CODE_UNKOWN)
		resp.ErrDesc = new(string)
		*resp.ErrDesc = "get player failed"
		return
	}
}
