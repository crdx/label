package main

import (
	"fmt"
	"os"
	"strings"

	"crdx.org/col"
	"crdx.org/duckopt/v2"
	"crdx.org/label/pkg/driver"
	"crdx.org/label/pkg/kitty"
	"crdx.org/label/pkg/ptouch"
	"crdx.org/label/pkg/render"
	"crdx.org/logger"
)

const glyphHeight = 50

func getUsage() string {
	return `
		Usage:
			$0 [options] print <text>
			$0 [options] status

		Options:
			--chain        Don't cut the label
			-p, --preview  Preview the label in the terminal instead of printing it
	`
}

type Opts struct {
	Print   bool   `docopt:"print"`
	Status  bool   `docopt:"status"`
	Text    string `docopt:"<text>"`
	Chain   bool   `docopt:"--chain"`
	Preview bool   `docopt:"--preview"`
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
	if opts.Print && strings.TrimSpace(opts.Text) == "" {
		return fmt.Errorf("text must not be empty")
	}

	if opts.Print && opts.Preview {
		return previewLabel(opts)
	}

	printer, err := ptouch.Open()
	if err != nil {
		return err
	}
	defer printer.Close()

	switch {
	case opts.Print:
		return printLabel(printer, opts)
	case opts.Status:
		return showStatus(printer)
	}
	return fmt.Errorf("no command given")
}

func showStatus(printer ptouch.Printer) error {
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

	tapeLength := "Continuous"
	if status.TapeLength != 0 {
		tapeLength = fmt.Sprintf("%dmm", status.TapeLength)
	}
	row("Length", tapeLength)

	if status.TapeWidth != 0 {
		row("Colour", fmt.Sprintf("%s on %s", status.FontColor, status.TapeColor))
	}

	return nil
}

func printLabel(printer ptouch.Printer, opts Opts) error {
	status, err := printer.Status()
	if err != nil {
		return err
	}

	if statusErrors := status.Errors(); len(statusErrors) > 0 {
		return fmt.Errorf("printer reports: %s", strings.Join(statusErrors, ", "))
	}

	if status.TapeWidth == 0 {
		return fmt.Errorf("no tape loaded")
	}

	printableDots, err := status.PrintableDots()
	if err != nil {
		return err
	}

	img, err := render.Text(opts.Text, min(glyphHeight, printableDots), ptouch.PrintHeadWidth)
	if err != nil {
		return fmt.Errorf("render text: %w", err)
	}

	data, bytesWidth, err := ptouch.Rasterise(img)
	if err != nil {
		return err
	}

	packed, err := ptouch.Pack(data, bytesWidth)
	if err != nil {
		return err
	}

	logger.Info("printing on %dmm %s tape", status.TapeWidth, status.MediaType)

	return printer.Print(packed, len(data)/bytesWidth, status.TapeWidth, !opts.Chain)
}

func previewLabel(opts Opts) error {
	img, err := render.Text(opts.Text, min(glyphHeight, ptouch.PrintHeadWidth), ptouch.PrintHeadWidth)
	if err != nil {
		return fmt.Errorf("render text: %w", err)
	}

	return kitty.PrintImage(os.Stdout, img)
}

func row(name string, value string) {
	fmt.Printf("%s %s\n", col.Dim("%-11s", name+":"), value)
}
