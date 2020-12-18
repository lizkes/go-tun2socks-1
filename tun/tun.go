package tun

import (
	"io"

	"golang.zx2c4.com/wireguard/tun"
)

func OpenTunDevice(name, addr, gw, mask string, dnsServers []string, persist bool, mtu int) (io.ReadWriteCloser, error) {
	tunDev, err := tun.CreateTUN(name, mtu)
	if err != nil {
		return nil, err
	}

	getName, err := tunDev.Name()
	if err != nil {
		return nil, err
	}

	return &tunnel{Device: tunDev}, setInterface(getName, addr, gw, mask, mtu, tunDev.(*tun.NativeTun))
}
