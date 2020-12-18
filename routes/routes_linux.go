package routes

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

func routeAdd(dst interface{}, gw net.IP, priority int, name string) error {
	route := netlink.Route{
		Dst:      getNet(dst),
		Priority: priority,
		Gw:       gw,
	}
	if gw == nil {
		link, err := netlink.LinkByName(name)
		if err != nil {
			return fmt.Errorf("failed to get %q interface by name: %s", name, err)
		}
		route.LinkIndex = link.Attrs().Index
	}
	if err := netlink.RouteReplace(&route); err != nil {
		return fmt.Errorf("failed to add %s route to %q interface: %s", dst, name, err)
	}
	return nil
}

func routeDel(dst interface{}, gw net.IP, priority int, name string) error {
	route := netlink.Route{
		Dst:      getNet(dst),
		Priority: priority,
		Gw:       gw,
	}
	if gw == nil {
		link, err := netlink.LinkByName(name)
		if err != nil {
			return fmt.Errorf("failed to get %q interface by name: %s", name, err)
		}
		route.LinkIndex = link.Attrs().Index
	}
	if err := netlink.RouteDel(&route); err != nil {
		return fmt.Errorf("failed to delete %s route from %q interface: %s", dst, name, err)
	}
	return nil
}
