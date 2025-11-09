package main

import (
	"context"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/vishvananda/netlink"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"nuist_rover/configuration"
	"nuist_rover/logger"
	"nuist_rover/nuistnet"
	"nuist_rover/nuistnet/model"
	"net"
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

	if config.RetryInterval <= 0 {
		config.RetryInterval = 30 * time.Second
		log.Info("retry interval has empty value, defaulting to %s", config.RetryInterval.String())
	}

	log.Log("loaded %d account(s)", len(config.Accounts))
	ctx, cancelCtx := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
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
		done := make(chan struct{})
		go func() {
			dial_all_parallel(ctx, *config, log)
			close(done)
		}()
		select {
		case <-done:
		case sig := <-signals:
			cancelCtx()
			log.Log("%s", sig.String())
			<-done
		}
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

	if config.CheckOnlineViaPortal {
		method := config.OnlineCheck.Method
		if method == "" {
			method = "portal"
		}

		var signedIn bool
		var err error

		if method == "portal" {
			onlineCheckCtx, cancelOnlineCheckCtx := context.WithTimeout(ctx, 10*time.Second)
			defer cancelOnlineCheckCtx()

			signedIn, err = client.IsOnline(onlineCheckCtx)
			if err != nil {
				log.Warning("cannot query dial state via portal: %s", err)
			}
		} else if method == "ping" {
			host := config.OnlineCheck.Host
			if host == "" {
				host = "8.8.8.8"
			}
			count := config.OnlineCheck.Count
			if count <= 0 {
				count = 3
			}
			threshold := config.OnlineCheck.Threshold
			if threshold <= 0 {
				threshold = 0.5
			}

			signedIn, err = isOnlineViaPing(host, count, threshold)
			if err != nil {
				log.Warning("cannot query dial state via ping: %s", err)
			}
		} else {
			log.Warning("unknown online check method: %s", method)
		}

		if err == nil && signedIn {
			log.Info("already online on %s", nic)
			return
		}
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
			if remainingTrails > 0 {
				log.Log("waiting %s before next retry", config.RetryInterval.String())
				time.Sleep(config.RetryInterval)
			}
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

func isOnlineViaPing(host string, count int, threshold float64) (bool, error) {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return false, err
	}
	defer c.Close()

	successCount := 0
	for i := 0; i < count; i++ {
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID:   os.Getpid() & 0xffff,
				Seq:  i,
				Data: []byte("ping"),
			},
		}
		wb, err := wm.Marshal(nil)
		if err != nil {
			return false, err
		}

		dst, err := net.ResolveIPAddr("ip4", host)
		if err != nil {
			return false, err
		}

		if _, err := c.WriteTo(wb, dst); err != nil {
			return false, err
		}

		rb := make([]byte, 1500)
		c.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, _, err := c.ReadFrom(rb)
		if err != nil {
			continue // timeout or error, consider failed
		}

		rm, err := icmp.ParseMessage(1, rb[:n])
		if err != nil {
			continue
		}

		if rm.Type == ipv4.ICMPTypeEchoReply {
			successCount++
		}
	}

	successRate := float64(successCount) / float64(count)
	return successRate >= threshold, nil
}
