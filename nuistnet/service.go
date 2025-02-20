package nuistnet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"nuist_rover/nuistnet/helper"
	"nuist_rover/nuistnet/isp"
	"nuist_rover/nuistnet/model"
	"strconv"
	"sync"
)

func (c Client) GetApiV1() string {
	return fmt.Sprintf("%s/api/v1/login", c.ServerUrl)
}

func (c Client) GetIspMapping(account model.Account) (map[isp.Type]int, error) {
	req := model.GetReqModelBase(account)
	req.Channel = "_GET"
	req.Pagesign = "firstauth"

	for addr, client := range c.clients {
		req.UsrIpAdd = addr.(*net.TCPAddr).IP.String()
		body, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}

		response, err := client.Post(c.GetApiV1(), "application/json", bytes.NewBuffer(body))
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

func (c Client) Signin(account model.Account) (map[string]model.SigninContent, error) {
	ispMapping, err := c.GetIspMapping(account)
	if err != nil {
		return nil, err
	}

	var errorList []error
	responseMap := make(map[string]model.SigninContent)
	var wg = sync.WaitGroup{}
	var errorListLock sync.Mutex
	var responseListLock sync.Mutex
	wg.Add(len(c.clients))

	for addr, httpClient := range c.clients {
		go func() {
			defer wg.Done()

			req := model.GetReqModel(account, ispMapping)
			req.Pagesign = "secondauth"
			req.UsrIpAdd = addr.(*net.TCPAddr).IP.String()

			body, err := json.Marshal(req)
			if err != nil {
				panic(err)
			}
			response, err := httpClient.Post(c.GetApiV1(), "application/json", bytes.NewBuffer(body))
			if err != nil {
				errorListLock.Lock()
				errorList = append(errorList, errors.New(fmt.Sprintf("could not connect to authentication server: %s", err)))
				errorListLock.Unlock()
				return
			}
			buffer, err := io.ReadAll(helper.GetBody(response))
			if err != nil {
				errorList = append(errorList, err)
			}
			var responseBodyBase model.Response[any]
			err = json.Unmarshal(buffer, &responseBodyBase)
			if err != nil {
				errorListLock.Lock()
				errorList = append(errorList, err)
				errorListLock.Unlock()
				return
			}

			if responseBodyBase.Code != 200 {
				errorListLock.Lock()
				errorList = append(
					errorList,
					errors.New(fmt.Sprintf("failure response code %d from authenication server", responseBodyBase.Code)),
				)
				errorListLock.Unlock()
				return
			}

			var responseBody model.Response[model.SigninContent]

			err = json.Unmarshal(buffer, &responseBody)
			if err != nil {
				errorList = append(errorList, err)
				return
			}

			responseListLock.Lock()
			responseMap[c.NicInterface.Name] = responseBody.Data
			responseListLock.Unlock()
		}()
	}

	wg.Wait()

	if len(errorList) > 0 {
		err = errors.Join(errorList...)
	}
	return responseMap, err
}
