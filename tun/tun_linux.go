package tun

import (
	"fmt"

	"github.com/eycorsican/go-tun2socks/routes"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/tun"
)

func setInterface(name, addr, gw, mask string, tun *tun.NativeTun) error {
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
