package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"nuist_rover/configuration"
	"nuist_rover/logger"
	"nuist_rover/nuistnet"
	"sync"
)

var args struct {
	Configuration string `short:"c" optional:"" help:"Name of the configuration file." type:"file"`
	Retry         bool
	Verbose       string `enum:"log,info,warning,exception,unknown" default:"unknown"`
}

func parseLogLevel(args ...string) logger.LogLevel {
	for _, item := range args {
		if len(item) <= 0 || item == "unknown" || len(item) <= 0 {
			continue
		}

		parsed := logger.ParseLevel(item)
		if parsed != logger.UNKNOWN {
			return parsed
		}
	}

	if len(args) > 0 {
		fmt.Printf("unknown log level: %s\n", args[len(args)-1])
	}

	return logger.UNKNOWN
}

func main() {
	kong.Parse(&args)
	if len(args.Configuration) <= 0 {
		args.Configuration = "/etc/nuistrover/config.toml"
	}

	config, err := configuration.Parse(args.Configuration)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var log logger.Logger
	log.Level = parseLogLevel(args.Verbose, config.Verbose)

	if config.Retry > 0 || args.Retry {
		config.Retry = max(config.Retry, 1)
	}

	if len(config.ServerUrl) <= 0 {
		fmt.Println("warning: server url has empty value")
	}

	var wg sync.WaitGroup
	wg.Add(len(config.Accounts))

	log.Log("loaded %d accounts", len(config.Accounts))
	for nic, account := range config.Accounts {
		go func() {
			remainingTrails := config.Retry + 1
			for remainingTrails > 0 {
				client, err := nuistnet.NewClient(config.ServerUrl, nic)
				if err != nil {
					panic(err)
				}
				responses, err := client.Signin(account)
				successful := len(responses) > 0
				if err != nil {
					var level logger.LogLevel
					if !successful {
						level = logger.EXCEPTION
					} else {
						level = logger.WARNING
					}
					log.Println(level, "failed to dial via %s using %s: %s", nic, account.Username, err)
				}

				if !successful {
					remainingTrails -= 1
				} else {
					log.Info("dial succeeded on %s", nic)
					break
				}
			}

			wg.Done()
		}()
	}

	wg.Wait()
}
