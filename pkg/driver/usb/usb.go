package usb

import (
	"fmt"
	"io"

	"github.com/google/gousb"
)

const (
	brotherVendorID = 0x04f9
	productID       = 0x20af // PT-P710BT
)

type USB struct {
	input  *gousb.InEndpoint
	output *gousb.OutEndpoint
	done   func()
}

func Open() (io.ReadWriteCloser, error) {
	ctx := gousb.NewContext()

	success := false
	defer func() {
		if !success {
			_ = ctx.Close()
		}
	}()

	device, err := ctx.OpenDeviceWithVIDPID(brotherVendorID, productID)
	if err != nil {
		return nil, fmt.Errorf("open USB device: %w", err)
	}
	if device == nil {
		return nil, fmt.Errorf("USB device not found")
	}
	defer func() {
		if !success {
			_ = device.Close()
		}
	}()

	if err := device.SetAutoDetach(true); err != nil {
		return nil, fmt.Errorf("set auto detach kernel driver: %w", err)
	}

	usbInterface, done, err := device.DefaultInterface()
	if err != nil {
		return nil, fmt.Errorf("get default interface: %w", err)
	}
	defer func() {
		if !success {
			done()
		}
	}()

	input, err := usbInterface.InEndpoint(0x81)
	if err != nil {
		return nil, fmt.Errorf("open InEndpoint: %w", err)
	}

	output, err := usbInterface.OutEndpoint(0x02)
	if err != nil {
		return nil, fmt.Errorf("open OutEndpoint: %w", err)
	}

	success = true
	return &USB{
		input:  input,
		output: output,
		done: func() {
			done()
			_ = device.Close()
			_ = ctx.Close()
		},
	}, nil
}

func (self *USB) Close() error {
	self.done()
	return nil
}

func (self *USB) Write(b []byte) (int, error) {
	return self.output.Write(b)
}

func (self *USB) Read(b []byte) (int, error) {
	return self.input.Read(b)
}
