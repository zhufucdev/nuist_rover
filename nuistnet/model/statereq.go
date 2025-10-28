package model

import "nuist_rover/nuistnet/encryption"

type NusitNetOnlineStateQueryReq struct {
	GetUserOnlineState string `json:"getuseronlinestate"`
	UsrIpAdd           string `json:"user_ipadress"`
}

func (req NusitNetOnlineStateQueryReq) Encrypt() NusitNetOnlineStateQueryReq {
	encState, _ := encryption.Encrypt(encryption.ENCRYPTION_KEY, "on_or_off")
	encIp, _ := encryption.Encrypt(encryption.ENCRYPTION_KEY, req.UsrIpAdd)
	return NusitNetOnlineStateQueryReq{
		GetUserOnlineState: encState,
		UsrIpAdd:           encIp,
	}
}
