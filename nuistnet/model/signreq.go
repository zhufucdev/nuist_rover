package model

import (
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
