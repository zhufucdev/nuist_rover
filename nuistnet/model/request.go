package model

import (
	"nuist_rover/nuistnet/isp"
	"strconv"
)

type NuistNetReq struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	IfAutoLogin string `json:"ifautologin"`
	Channel     string `json:"channel"`
	Pagesign    string `json:"pagesign"`
	UsrIpAdd    string `json:"usripadd"`
}

func GetReqModel(account Account, ispMapping map[isp.Type]int) NuistNetReq {
	base := GetReqModelBase(account)
	base.Channel = strconv.Itoa(ispMapping[account.Isp])
	return base
}

func GetReqModelBase(account Account) NuistNetReq {
	return NuistNetReq{
		Username:    account.Username,
		Password:    account.Password,
		IfAutoLogin: "0",
	}
}
