//go:build !linux

package bt

import (
	"errors"
	"io"
)

func Open() (io.ReadWriteCloser, string, error) {
	return nil, "", errors.New("bluetooth is only supported on linux")
}
