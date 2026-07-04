package driver

import (
	"fmt"
	"io"
	"strings"

	"crdx.org/label/pkg/driver/bt"
	"crdx.org/label/pkg/driver/usb"
)

type Transport string

const (
	TransportUSB Transport = "usb"
	TransportBT  Transport = "bt"
)

type Connection struct {
	io.ReadWriteCloser

	Transport Transport
	Address   string
}

func Open() (Connection, error) {
	var failures []string

	usbConnection, err := usb.Open()
	if err == nil {
		return Connection{
			ReadWriteCloser: usbConnection,
			Transport:       TransportUSB,
		}, nil
	}
	failures = append(failures, fmt.Sprintf("%s: %v", TransportUSB, err))

	btConnection, address, err := bt.Open()
	if err == nil {
		return Connection{
			ReadWriteCloser: btConnection,
			Transport:       TransportBT,
			Address:         address,
		}, nil
	}
	failures = append(failures, fmt.Sprintf("%s: %v", TransportBT, err))

	return Connection{}, fmt.Errorf("no printer found (%s)", strings.Join(failures, "; "))
}
