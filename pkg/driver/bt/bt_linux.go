package bt

import (
	"fmt"
	"io"
	"net"
	"os"
	"slices"

	"golang.org/x/sys/unix"
)

func Open() (io.ReadWriteCloser, string, error) {
	address, err := discover()
	if err != nil {
		return nil, "", fmt.Errorf("discover bluetooth printer: %w", err)
	}

	mac, err := parseMAC(address)
	if err != nil {
		return nil, "", err
	}

	fd, err := unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, unix.BTPROTO_RFCOMM)
	if err != nil {
		return nil, "", fmt.Errorf("open RFCOMM socket: %w", err)
	}

	if err := unix.Connect(fd, &unix.SockaddrRFCOMM{Addr: mac, Channel: 1}); err != nil {
		_ = unix.Close(fd)
		return nil, "", fmt.Errorf("connect to %s: %w", address, err)
	}

	return os.NewFile(uintptr(fd), address), address, nil
}

func parseMAC(address string) ([6]uint8, error) {
	hardwareAddress, err := net.ParseMAC(address)
	if err != nil || len(hardwareAddress) != 6 {
		return [6]uint8{}, fmt.Errorf("invalid bluetooth address %q", address)
	}
	slices.Reverse(hardwareAddress)
	return [6]uint8(hardwareAddress), nil
}
