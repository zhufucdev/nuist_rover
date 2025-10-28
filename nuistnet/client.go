package nuistnet

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
)

type Client struct {
	ServerUrl    string
	NicInterface net.Interface
	clients      map[net.Addr]http.Client
}

func NewClient(serverUrl string, nicName string) (Client, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return Client{}, err
	}
	var targetNic *net.Interface = nil
	for _, nic := range interfaces {
		if nic.Name == nicName {
			targetNic = &nic
			break
		}
	}
	if targetNic == nil {
		return Client{}, fmt.Errorf("network interface named %s was not found", nicName)
	}

	addresses, err := targetNic.Addrs()
	if err != nil {
		return Client{}, err
	}
	client := make(map[net.Addr]http.Client)
	for _, addr := range addresses {
		localAddr, err := getTcpAddr(addr)
		if err != nil || localAddr.IP.IsLinkLocalUnicast() || localAddr.IP.IsLinkLocalMulticast() || localAddr.IP.To4() == nil {
			continue
		}
		dialer := net.Dialer{LocalAddr: localAddr}
		client[localAddr] = http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}}
	}

	return Client{
		ServerUrl:    serverUrl,
		NicInterface: *targetNic,
		clients:      client,
	}, err
}

func getTcpAddr(addr net.Addr) (*net.TCPAddr, error) {
	switch addr := addr.(type) {
	case *net.IPNet:
		ipAddr := addr
		ip4 := ipAddr.IP.To4()
		var netipAddr netip.Addr
		if ip4 != nil {
			netipAddr = netip.AddrFrom4([4]byte(ip4))
		} else {
			ip16 := ipAddr.IP.To16()
			if ip16 != nil {
				netipAddr = netip.AddrFrom16([16]byte(ip16))
			} else {
				return nil, fmt.Errorf("unknown ip length %d", len(ipAddr.IP))
			}
		}

		return net.TCPAddrFromAddrPort(netip.AddrPortFrom(netipAddr, 0)), nil
	}
	return nil, errors.New("unknown address type")
}
