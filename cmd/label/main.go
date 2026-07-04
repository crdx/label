package main

import (
	"fmt"

	"crdx.org/col"
	"crdx.org/duckopt/v2"
	"crdx.org/label/pkg/driver"
	"crdx.org/label/pkg/ptouch"
	"crdx.org/logger"
)

func getUsage() string {
	return `
		Usage:
			$0 [options] status
	`
}

type Opts struct {
	Status bool `docopt:"status"`
}

func main() {
	logger.InitStderr()

	opts := duckopt.MustBind[Opts](getUsage(), "$0")
	col.Init()

	if err := run(*opts); err != nil {
		logger.Fatal(err)
	}
}

func run(opts Opts) error {
	printer, err := ptouch.Open()
	if err != nil {
		return err
	}
	defer printer.Close()

	switch {
	case opts.Status:
		return printStatus(printer)
	default:
		return fmt.Errorf("no command given")
	}
}

func printStatus(printer ptouch.Printer) error {
	status, err := printer.Status()
	if err != nil {
		return err
	}

	tape := col.Dim("none")
	if status.TapeWidth != 0 {
		tape = fmt.Sprintf("%s, %dmm", status.MediaType, status.TapeWidth)
	}

	row("Printer", status.Model.String())

	connection := string(printer.Driver.Transport)
	if printer.Driver.Transport == driver.TransportBT {
		connection = "bluetooth"
	}
	if printer.Driver.Address != "" {
		connection = fmt.Sprintf("%s (%s)", connection, printer.Driver.Address)
	}
	row("Connection", connection)

	row("Power", fmt.Sprint(status.Battery))
	row("Tape", tape)

	tapeLength := "continuous"
	if status.TapeLength != 0 {
		tapeLength = fmt.Sprintf("%dmm", status.TapeLength)
	}
	row("Length", tapeLength)

	if status.TapeWidth != 0 {
		row("Colour", fmt.Sprintf("%s on %s", status.FontColor, status.TapeColor))
	}

	return nil
}

func row(name string, value string) {
	fmt.Printf("%s %s\n", col.Dim("%-11s", name+":"), value)
}
