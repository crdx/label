package main

import (
	"fmt"
	"strconv"
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
			$0 print [--chain] <text>
			$0 preview [-m <mm>] <text>
			$0 status

		Options:
			--chain          Don't cut the label
			-m, --mm <mm>    Tape width in mm to preview for (default: loaded tape)
	`
}

type Opts struct {
	Print   bool   `docopt:"print"`
	Preview bool   `docopt:"preview"`
	Status  bool   `docopt:"status"`
	Text    string `docopt:"<text>"`
	Chain   bool   `docopt:"--chain"`
	MM      string `docopt:"--mm"`
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
	if (opts.Print || opts.Preview) && strings.TrimSpace(opts.Text) == "" {
		return fmt.Errorf("text must not be empty")
	}

	if opts.Preview {
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
	width, err := resolveTapeWidth(opts.MM)
	if err != nil {
		return err
	}

	printableDots, err := ptouch.PrintableDots(width)
	if err != nil {
		return err
	}

	tapeDots, err := ptouch.TapeDots(width)
	if err != nil {
		return err
	}

	img, err := render.Text(opts.Text, min(glyphHeight, printableDots), tapeDots)
	if err != nil {
		return fmt.Errorf("render text: %w", err)
	}

	return kitty.PrintImage(img)
}

func resolveTapeWidth(mm string) (ptouch.TapeWidth, error) {
	if mm == "" {
		width, err := ptouch.LoadedTapeWidth()
		if err != nil {
			return 0, fmt.Errorf("%w (pass -m to preview without a printer)", err)
		}
		return width, nil
	}

	width, err := strconv.Atoi(mm)
	if err != nil {
		return 0, fmt.Errorf("invalid tape width %q: must be a number in mm", mm)
	}
	return ptouch.TapeWidth(width), nil
}

func row(name string, value string) {
	fmt.Printf("%s %s\n", col.Dim("%-11s", name+":"), value)
}
