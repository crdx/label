package bt

import (
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
)

const (
	deviceInterface   = "org.bluez.Device1"
	printerNamePrefix = "PT-P710BT"
)

type pairedPrinter struct {
	name    string
	address string
}

func discover() (string, error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return "", fmt.Errorf("connect to system bus (is D-Bus running?): %w", err)
	}
	defer func() { _ = conn.Close() }()

	var objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err = conn.Object("org.bluez", "/").
		Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).
		Store(&objects)
	if err != nil {
		return "", fmt.Errorf("enumerate bluetooth devices: %w", err)
	}

	var printers []pairedPrinter
	for _, interfaces := range objects {
		properties, found := interfaces[deviceInterface]
		if !found {
			continue
		}

		name, _ := properties["Name"].Value().(string)
		address, _ := properties["Address"].Value().(string)
		paired, _ := properties["Paired"].Value().(bool)

		if paired && address != "" && strings.HasPrefix(name, printerNamePrefix) {
			printers = append(printers, pairedPrinter{name: name, address: address})
		}
	}

	switch len(printers) {
	case 0:
		return "", fmt.Errorf("no paired P-touch printer found (pair one with bluetoothctl)")
	case 1:
		return printers[0].address, nil
	default:
		var descriptions []string
		for _, printer := range printers {
			descriptions = append(descriptions, fmt.Sprintf("%s (%s)", printer.name, printer.address))
		}
		return "", fmt.Errorf(
			"multiple paired P-touch printers found (%s); need exactly one",
			strings.Join(descriptions, ", "),
		)
	}
}
