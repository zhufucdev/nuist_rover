package configuration

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"nuist_rover/nuistnet/isp"
	"nuist_rover/nuistnet/model"
	"strings"
	"time"
)

func (r root) toRoot() Root {
	accounts := make(map[string]model.Account)
	for nic, acc := range r.Accounts {
		accounts[nic] = model.Account{
			Username: acc.Username,
			Password: acc.Password,
			Isp:      isp.Parse(acc.Isp),
		}
	}
	testInterval, err := time.ParseDuration(r.TestInterval)
	if err != nil {
		testInterval = 0
	}
	serverUrl := r.ServerUrl
	if !strings.HasPrefix(serverUrl, "http://") && !strings.HasPrefix(serverUrl, "https://") {
		serverUrl = "http://" + serverUrl
	}
	return Root{
		ServerUrl:    serverUrl,
		Retry:        r.Retry,
		TestInterval: testInterval,
		Verbose:      r.Verbose,
		RestartLink:  r.RestartLink,
		Accounts:     accounts,
	}
}

func Parse(filename string) (*Root, error) {
	var config root
	_, err := toml.DecodeFile(filename, &config)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %s", err)
	}

	rootConfig := config.toRoot()
	return &rootConfig, nil
}
