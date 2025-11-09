package configuration

import (
	"nuist_rover/nuistnet/model"
	"time"
)

type account struct {
	Username string
	Password string
	Isp      string
}

type OnlineCheck struct {
	Enabled   bool
	Method    string
	Host      string
	Count     int
	Threshold float64
}

type root struct {
	ServerUrl             string
	Retry                 uint
	RetryInterval         string
	TestInterval          string
	Verbose               string
	RestartLink           bool
	OnlineCheck           OnlineCheck
	Accounts              map[string]account
}

type Root struct {
	ServerUrl             string
	Retry                 uint
	RetryInterval         time.Duration
	TestInterval          time.Duration
	Verbose               string
	RestartLink           bool
	OnlineCheck           OnlineCheck
	Accounts              map[string]model.Account
}
