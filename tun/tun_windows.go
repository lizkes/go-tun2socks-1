package tun

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/eycorsican/go-tun2socks/routes"
	"golang.zx2c4.com/wireguard/tun"
)

func setInterface(name, addr, gw, mask string, tun *tun.NativeTun) error {
	addrs, err := routes.ParseAddresses(addr, gw, mask)
	if err != nil {
		return err
	}
	args := []string{
		"interface",
		"ipv4",
		"set",
		"address",
		"name=" + name,
		"static",
		addrs[0].IP.String(),
		net.IP(addrs[1].Mask).To4().String(),
	}
	v, err := exec.Command("netsh.exe", args...).Output()
	if err != nil {
		return fmt.Errorf("failed to set tun interface: %s: %s", v, err)
	}

	return nil
}
