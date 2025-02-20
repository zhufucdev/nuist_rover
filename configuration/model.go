package configuration

import "nuist_rover/nuistnet/model"

type account struct {
	Username string
	Password string
	Isp      string
}

type root struct {
	ServerUrl string
	Retry     uint
	Verbose   string
	Accounts  map[string]account
}

type Root struct {
	ServerUrl string
	Retry     uint
	Verbose   string
	Accounts  map[string]model.Account
}
