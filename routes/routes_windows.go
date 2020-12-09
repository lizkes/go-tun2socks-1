package routes

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
)

var re = regexp.MustCompile(`IfIndex\s+:\s+(\d+)`)
var ifIndex string

func getIfID(iface string) error {
	args := []string{
		"int",
		"ipv4",
		"show",
		"interfaces",
		iface,
	}
	v, err := exec.Command("netsh.exe", args...).Output()
	if err != nil {
		return fmt.Errorf("failed to find a %q interface index: %s, %s", iface, v, err)
	}
	if v := re.FindSubmatch(v); len(v) == 2 {
		ifIndex = string(v[1])
		return nil
	}

	return fmt.Errorf("failed to find a %q interface index: %s", iface, v)
}

func routeAdd(dst interface{}, gw net.IP, priority int, iface string) error {
	if ifIndex == "" {
		if err := getIfID(iface); err != nil {
			return err
		}
	}

	// an implementation of "replace"
	routeDel(dst, gw, priority, iface)
	d := getNet(dst)
	args := []string{
		"add",
		d.IP.String(),
		"mask",
		net.IP(d.Mask).To4().String(),
		gw.String(),
		"metric",
		fmt.Sprintf("%d", priority+1),
		"if",
		ifIndex,
	}
	v, err := exec.Command("route.exe", args...).Output()
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
		"mask",
		net.IP(d.Mask).To4().String(),
		gw.String(),
		"metric",
		fmt.Sprintf("%d", priority+1),
		"if",
		ifIndex,
	}
	v, err := exec.Command("route.exe", args...).Output()
	if err != nil {
		return fmt.Errorf("failed to delete %s route from %s interface: %s: %s", dst, iface, v, err)
	}
	return nil
}
