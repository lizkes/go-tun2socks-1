package routes

import (
	"fmt"
	"net"
	"os/exec"
)

func routeAdd(dst interface{}, gw net.IP, priority int, iface string) error {
	// an implementation of "replace"
	routeDel(dst, gw, priority, iface)
	d := getNet(dst)
	args := []string{
		"add",
		d.IP.String(),
		net.IP(d.Mask).To4().String(),
		gw.String(),
		"metric",
		fmt.Sprintf("%d", priority),
	}
	v, err := exec.Command("route", args...).Output()
	if err != nil {
		return fmt.Errorf("failed to add %s route to %s interface: %s: %s", dst, iface, v, err)
	}
	return nil
}

func routeDel(dst interface{}, gw net.IP, priority int, iface string) error {
	d := getNet(dst)
	args := []string{
		"delete",
		d.IP.String(),
		net.IP(d.Mask).To4().String(),
		gw.String(),
		"metric",
		fmt.Sprintf("%d", priority),
	}
	v, err := exec.Command("route", args...).Output()
	if err != nil {
		return fmt.Errorf("failed to delete %s route from %s interface: %s: %s", dst, iface, v, err)
	}
	return nil
}
