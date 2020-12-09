package tun

import (
	"io"
	"runtime"

	"golang.zx2c4.com/wireguard/tun"
)

type tunnel struct {
	tun.Device
}

func (t *tunnel) Read(b []byte) (int, error) {
	if runtime.GOOS == "windows" {
		return t.Device.Read(b, 0)
	}
	// unix.IFF_NO_PI is not set, therefore we receive packet information
	n, err := t.Device.File().Read(b)
	if n < 4 {
		return 0, err
	}
	// shift slice to the left
	return copy(b[:n-4], b[4:n]), nil
}

func (t *tunnel) Write(b []byte) (int, error) {
	if runtime.GOOS == "windows" {
		return t.Device.Write(b, 0)
	}
	return t.Device.Write(append(make([]byte, 4), b...), 4)
}

func (t *tunnel) Close() error {
	return t.Device.Close()
}

func OpenTunDevice(name, addr, gw, mask string, dnsServers []string, persist bool) (io.ReadWriteCloser, error) {
	tunDev, err := tun.CreateTUN(name, 1500)
	if err != nil {
		return nil, err
	}

	getName, err := tunDev.Name()
	if err != nil {
		return nil, err
	}

	return &tunnel{Device: tunDev}, setInterface(getName, addr, gw, mask, tunDev.(*tun.NativeTun))
}
