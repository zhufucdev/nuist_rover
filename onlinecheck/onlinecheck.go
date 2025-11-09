package onlinecheck

import (
	"context"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"nuist_rover/configuration"
	"nuist_rover/logger"
	"nuist_rover/nuistnet"
	"os"
	"time"
)

// CheckOnline performs online check based on configuration
func CheckOnline(ctx context.Context, config configuration.Root, nic string, client nuistnet.Client, log logger.Logger) (bool, error) {
	if !config.OnlineCheck.Enabled {
		return false, nil // not enabled, proceed with signin
	}

	method := config.OnlineCheck.Method
	if method == "" {
		method = "portal"
	}

	switch method {
	case "portal":
		return checkOnlineViaPortal(ctx, client, log)
	case "ping":
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
		return checkOnlineViaPing(host, count, threshold, log)
	default:
		log.Warning("unknown online check method: %s", method)
		return false, nil
	}
}

func checkOnlineViaPortal(ctx context.Context, client nuistnet.Client, log logger.Logger) (bool, error) {
	onlineCheckCtx, cancelOnlineCheckCtx := context.WithTimeout(ctx, 10*time.Second)
	defer cancelOnlineCheckCtx()

	signedIn, err := client.IsOnline(onlineCheckCtx)
	if err != nil {
		log.Warning("cannot query dial state via portal: %s", err)
		return false, err
	}
	return signedIn, nil
}

func checkOnlineViaPing(host string, count int, threshold float64, log logger.Logger) (bool, error) {
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