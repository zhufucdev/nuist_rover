package nuistnet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"nuist_rover/nuistnet/helper"
	"nuist_rover/nuistnet/isp"
	"nuist_rover/nuistnet/model"
	"strconv"
	"sync"
)

func (c Client) GetIspMapping(account model.Account) (map[isp.Type]int, error) {
	req := model.GetSignReqModelBase(account)
	req.Channel = "_GET"
	req.Pagesign = "firstauth"

	for addr, client := range c.clients {
		req.UsrIpAdd = addr.(*net.TCPAddr).IP.String()
		body, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}

		response, err := client.Post(loginApiV1(c.ServerUrl), "application/json", bytes.NewBuffer(body))
		if err != nil {
			continue
		}

		jsonDecoder := json.NewDecoder(helper.GetBody(response))
		var responseBody model.Response[model.ListChannelsContent]
		err = jsonDecoder.Decode(&responseBody)
		if err != nil || responseBody.Code != 200 {
			continue
		}

		mapping := make(map[isp.Type]int, len(responseBody.Data.Channels))
		for _, channel := range responseBody.Data.Channels {
			id, err := strconv.Atoi(channel.Id)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("error parsing channel ID from response (raw: %s): %s", channel.Id, err))
			}
			mapping[isp.Parse(channel.Name)] = id
		}
		return mapping, nil
	}

	return nil, errors.New("no usable network interface bounded")
}

func (c Client) Signin(account model.Account) (map[net.Addr]model.SigninContent, error) {
	return c.SigninWithContext(account, context.TODO())
}

func (c Client) SigninWithContext(account model.Account, ctx context.Context) (map[net.Addr]model.SigninContent, error) {
	ispMapping, err := c.GetIspMapping(account)
	if err != nil {
		return nil, err
	}

	return multicastRequestFull[model.SigninContent](func(addr net.Addr, client http.Client) any {
		req := model.GetSignReqModel(account, ispMapping)
		req.Pagesign = "secondauth"
		req.UsrIpAdd = addr.(*net.TCPAddr).IP.String()
		return req
	}, loginApiV1(c.ServerUrl), c.clients, ctx)
}

func (c Client) IsOnline() (bool, error) {
	data, err := multicastRequestFast[model.StateQueryContent](func(addr net.Addr, client http.Client) any {
		return model.NusitNetOnlineStateQueryReq{
			GetUserOnlineState: "on_or_off",
			UsrIpAdd:           addr.(*net.TCPAddr).IP.String(),
		}
	}, preloginApiV1(c.ServerUrl), c.clients, context.TODO())
	if data != nil {
		switch data.OnlineState {
		case "on":
			return true, nil
		case "off":
			return false, nil
		default:
			return false, fmt.Errorf("responded with unknown online state %s", data.OnlineState)
		}
	} else {
		return false, err
	}
}

func loginApiV1(serverUrl string) string {
	return fmt.Sprintf("%s/api/v1/login", serverUrl)
}

func preloginApiV1(serverUrl string) string {
	return fmt.Sprintf("%s/api/v1/pre_login", serverUrl)
}

func jsonPost[Data any](client http.Client, requestModel any, httpEndpoint string, ctx context.Context) (*Data, error) {
	body, err := json.Marshal(requestModel)
	if err != nil {
		panic(err)
	}
	request, err := http.NewRequestWithContext(ctx, "POST", httpEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not connect to authentication server: %s", err))
	}
	buffer, err := io.ReadAll(helper.GetBody(response))
	if err != nil {
		return nil, err
	}
	var responseBodyBase model.Response[any]
	err = json.Unmarshal(buffer, &responseBodyBase)
	if err != nil {
		return nil, err
	}

	if responseBodyBase.Code != 200 {
		return nil, errors.New(fmt.Sprintf("failure response code %d from authenication server", responseBodyBase.Code))
	}

	var responseBody model.Response[Data]

	err = json.Unmarshal(buffer, &responseBody)
	if err != nil {
		return nil, err
	}

	return &responseBody.Data, nil
}

func multicastRequestFull[Data any](requestModel func(addr net.Addr, client http.Client) any, httpEndpoint string, clientsByAddress map[net.Addr]http.Client, ctx context.Context) (result map[net.Addr]Data, err error) {
	var wg sync.WaitGroup
	errorMap := make(map[net.Addr]error)
	result = make(map[net.Addr]Data)

	wg.Add(len(clientsByAddress))

	for addr, httpClient := range clientsByAddress {
		go func() {
			defer wg.Done()
			response, err := jsonPost[Data](httpClient, requestModel(addr, httpClient), httpEndpoint, ctx)
			if err != nil {
				errorMap[addr] = err
			} else {
				result[addr] = *response
			}
		}()
	}

	wg.Wait()
	err = model.NewAggregatedNicError(errorMap)
	return
}

func multicastRequestFast[Data any](requestModel func(addr net.Addr, client http.Client) any, httpEndpoint string, clientsByAddress map[net.Addr]http.Client, ctx context.Context) (result *Data, err error) {
	var wg sync.WaitGroup
	errorMap := make(map[net.Addr]error)
	cancelCtx, cancelFn := context.WithCancel(ctx)

	wg.Add(len(clientsByAddress))
	for addr, httpClient := range clientsByAddress {
		go func() {
			defer wg.Done()
			response, err := jsonPost[Data](httpClient, requestModel(addr, httpClient), httpEndpoint, cancelCtx)
			if err != nil {
				errorMap[addr] = err
			} else {
				cancelFn()
				result = response
			}
		}()
	}

	wg.Wait()
	cancelFn()
	err = model.NewAggregatedNicError(errorMap)
	return
}
