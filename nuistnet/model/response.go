package model

type Response[Content any] struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    Content `json:"data"`
}

type ListChannelsContent struct {
	Channels []struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
}

type SigninContent struct {
	Reauth        bool   `json:"reauth"`
	Username      string `json:"username"`
	Balance       string `json:"balance"`
	Duration      string `json:"duration"`
	Outport       string `json:"outport"`
	TotalTimespan string `json:"totaltimespan"`
	UsrIpAdd      string `json:"usripadd"`
}

type StateQueryContent struct {
	OnlineState   string `json:"useronlinestate"`
	UserName      string `json:"username"`
	Balance       string `json:"balance"`
	Duration      string `json:"duration"`
	Outport       string `json:"outport"`
	TotalTimeSpan string `json:"totaltimespan"`
	UsrIpAdd      string `json:"useripadd"`
}
