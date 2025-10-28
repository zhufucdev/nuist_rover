package model

import (
	"nuist_rover/nuistnet/encryption"
	"nuist_rover/nuistnet/isp"
	"strconv"
)

type NuistNetSignReq struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	IfAutoLogin string `json:"ifautologin"`
	Channel     string `json:"channel"`
	Pagesign    string `json:"pagesign"`
	UsrIpAdd    string `json:"usripadd"`
}

func GetSignReqModel(account Account, ispMapping map[isp.Type]int) NuistNetSignReq {
	base := GetSignReqModelBase(account)
	base.Channel = strconv.Itoa(ispMapping[account.Isp])
	return base
}

func GetSignReqModelBase(account Account) NuistNetSignReq {
	return NuistNetSignReq{
		Username:    account.Username,
		Password:    account.Password,
		IfAutoLogin: "0",
	}
}

func (req NuistNetSignReq) Encrypt() NuistNetSignReq {
	encUsername, _ := encryption.Encrypt(encryption.ENCRYPTION_KEY, req.Username)

	encKey := encryption.GenerateEncryptionKey(req.Username)

	encPassword, _ := encryption.Encrypt(encKey, req.Password)
	encIfAutoLogin, _ := encryption.Encrypt(encKey, req.IfAutoLogin)
	encChannel, _ := encryption.Encrypt(encKey, req.Channel)
	encPagesign, _ := encryption.Encrypt(encKey, req.Pagesign)
	encUsrIpAdd, _ := encryption.Encrypt(encKey, req.UsrIpAdd)

	return NuistNetSignReq{
		Username:    encUsername,
		Password:    encPassword,
		IfAutoLogin: encIfAutoLogin,
		Channel:     encChannel,
		Pagesign:    encPagesign,
		UsrIpAdd:    encUsrIpAdd,
	}
}
