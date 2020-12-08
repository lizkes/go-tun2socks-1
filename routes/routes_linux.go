package routes

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

func routeAdd(dst interface{}, gw net.IP, priority int, iface string) error {
	route := netlink.Route{
		Dst:      getNet(dst),
		Priority: priority,
		Gw:       gw,
	}
	if gw == nil {
		link, err := netlink.LinkByName(iface)
		if err != nil {
			return fmt.Errorf("failed to get %q interface by name: %s", iface, err)
		}
		route.LinkIndex = link.Attrs().Index
	}
	if err := netlink.RouteReplace(&route); err != nil {
		return fmt.Errorf("failed to add %s route to %q interface: %s", dst, iface, err)
	}
	return nil
}

func routeDel(dst interface{}, gw net.IP, priority int, iface string) error {
	route := netlink.Route{
		Dst:      getNet(dst),
		Priority: priority,
		Gw:       gw,
	}
	if gw == nil {
		link, err := netlink.LinkByName(iface)
		if err != nil {
			return fmt.Errorf("failed to get %q interface by name: %s", iface, err)
		}
		route.LinkIndex = link.Attrs().Index
	}
	if err := netlink.RouteDel(&route); err != nil {
		return fmt.Errorf("failed to delete %s route from %q interface: %s", dst, iface, err)
	}
	return nil
}
