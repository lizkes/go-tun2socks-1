// +build !windows

package tun

import (
	"golang.zx2c4.com/wireguard/tun"
)

type tunnel struct {
	tun.Device
}

func (t *tunnel) Read(b []byte) (int, error) {
	// unix.IFF_NO_PI is not set, therefore we receive packet information
	n, err := t.Device.File().Read(b)
	if n < 4 {
		return 0, err
	}
	// shift slice to the left
	return copy(b[:n-4], b[4:n]), nil
}

func (t *tunnel) Write(b []byte) (int, error) {
	return t.Device.Write(append(make([]byte, 4), b...), 4)
}

func (t *tunnel) Close() error {
	return t.Device.Close()
}
