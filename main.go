package main

import (
	"context"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/vishvananda/netlink"
	"nuist_rover/configuration"
	"nuist_rover/logger"
	"nuist_rover/nuistnet"
	"nuist_rover/nuistnet/model"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var args struct {
	Configuration string `short:"c" optional:"" help:"Name of the configuration file." type:"file"`
	Retry         bool
	Daemon        bool   `short:"D"`
	Verbose       string `enum:"log,info,warning,exception,unknown" default:"unknown"`
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
		log.Warning("server url has empty value")
	}

	if args.Daemon && config.TestInterval <= 0 {
		config.TestInterval = 1 * time.Minute
		log.Info("running in daemon mode while test interval has empty value, defaulting to %s", config.TestInterval.String())
	}

	log.Log("loaded %d account(s)", len(config.Accounts))
	ctx, cancelCtx := context.WithCancel(context.Background())
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	if args.Daemon {
		dial_all_parallel(ctx, *config, log)

		ticker := time.NewTicker(config.TestInterval)
		for {
			select {
			case <-ticker.C:
				go dial_all_parallel(ctx, *config, log)

			case sig := <-signals:
				cancelCtx()
				log.Log("%s", sig.String())
				return
			}
		}
	} else {
		dial_all_parallel(ctx, *config, log)
	}
}

func dial_all_parallel(ctx context.Context, config configuration.Root, log logger.Logger) {
	var wg sync.WaitGroup
	wg.Add(len(config.Accounts))
	for nic, account := range config.Accounts {
		go func() {
			defer wg.Done()
			dial(ctx, nic, account, config, log)
		}()
	}
	wg.Wait()
}

func dial(ctx context.Context, nic string, account model.Account, config configuration.Root, log logger.Logger) {
	remainingTrails := config.Retry + 1
	client, err := nuistnet.NewClient(config.ServerUrl, nic)
	if err != nil {
		panic(err)
	}

	onlineCheckCtx, cancelOnlineCheckCtx := context.WithTimeout(ctx, 10*time.Second)
	defer cancelOnlineCheckCtx()

	signedIn, err := client.IsOnline(onlineCheckCtx)
	if err != nil {
		log.Warning("cannot query dial state: %s", err)
	} else if signedIn {
		log.Info("already dialed on %s", nic)
		return
	}

	for remainingTrails > 0 {
		responses, err := client.SigninWithContext(account, ctx)
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
			log.Log("%d retrial(s) remaining", remainingTrails)
		} else {
			log.Info("dial succeeded on %s", nic)
			return
		}
	}

	if config.RestartLink {
		log.Info("retry expired, interface %s is restarting", nic)
		lo, err := netlink.LinkByName(nic)
		if err != nil {
			log.Error("%s was not found: %s", nic, err)
			return
		}
		err = netlink.LinkSetDown(lo)
		if err != nil {
			log.Error("failed to set down %s: %s", nic, err)
			return
		}
		err = netlink.LinkSetUp(lo)
		if err != nil {
			log.Error("failed to set up %s: %s", nic, err)
		}
	}
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
