package tun

import (
	"fmt"
	"io"

	"github.com/eycorsican/go-tun2socks/routes"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

func OpenTunDevice(name, addr, gw, mask string, dnsServers []string, persist bool) (io.ReadWriteCloser, error) {
	cfg := water.Config{
		DeviceType: water.TUN,
	}
	cfg.Name = name
	cfg.Persist = persist
	tunDev, err := water.New(cfg)
	if err != nil {
		return nil, err
	}
	name = tunDev.Name()

	return tunDev, setInterface(name, addr, gw, mask)
}

func setInterface(name, addr, gw, mask string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("failed to detect %s interface: %s", name, err)
	}

	addrs, err := routes.ParseAddresses(addr, gw, mask)
	if err != nil {
		return err
	}

	ipv4Addr := &netlink.Addr{
		IPNet: addrs[0],
		Peer:  addrs[1],
	}
	err = netlink.AddrAdd(link, ipv4Addr)
	if err != nil {
		return fmt.Errorf("failed to set peer address on %s interface: %s", name, err)
	}

	err = netlink.LinkSetUp(link)
	if err != nil {
		return fmt.Errorf("failed to set %s interface up: %s", name, err)
	}

	return nil
}
