package model

import (
	"nuist_rover/nuistnet/isp"
)

type Account struct {
	Username string
	Password string
	Isp      isp.Type
}
