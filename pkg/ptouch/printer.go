package ptouch

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"crdx.org/label/pkg/driver"
	"crdx.org/label/pkg/util"
)

type Driver struct {
	Transport driver.Transport
	Address   string
}

type Printer struct {
	Connection io.ReadWriteCloser
	Driver     Driver
}

type Status struct {
	Type       StatusType
	Model      Model
	Battery    BatteryStatusType
	Error1     Error1Type
	Error2     Error2Type
	MediaType  MediaType
	TapeColor  TapeColor
	TapeLength int
	TapeWidth  TapeWidth
	FontColor  FontColor
}

func Open() (Printer, error) {
	connection, err := driver.Open()
	if err != nil {
		return Printer{}, err
	}

	printer := Printer{
		Connection: connection,
		Driver:     Driver{connection.Transport, connection.Address},
	}

	if err := printer.reset(); err != nil {
		printer.Close()
		return Printer{}, err
	}

	return printer, nil
}

// Status requests the current printer status. ESC i S (reference.txt:793-812).
func (self *Printer) Status() (*Status, error) {
	if _, err := self.Connection.Write(cmdDumpStatus); err != nil {
		return nil, err
	}
	return self.readStatus()
}

func (self *Printer) Close() {
	_ = self.Connection.Close()
}

func LoadedTapeWidth() (TapeWidth, error) {
	printer, err := Open()
	if err != nil {
		return 0, err
	}
	defer printer.Close()

	status, err := printer.Status()
	if err != nil {
		return 0, err
	}

	return status.TapeWidth, nil
}

// Print runs the full print command sequence (reference.txt:170-181).
func (self *Printer) Print(data []byte, rasterLines int, width TapeWidth, cut bool) error {
	if err := self.setRasterMode(); err != nil {
		return err
	}
	if err := self.setNotificationMode(true); err != nil {
		return err
	}
	if err := self.setPrintProperty(rasterLines, width); err != nil {
		return err
	}
	if err := self.setPrintMode(cut, false); err != nil {
		return err
	}
	if err := self.setExtendedMode(cut, false, false, false); err != nil {
		return err
	}
	if err := self.setFeedAmount(10); err != nil {
		return err
	}
	if err := self.setCompressionModeEnabled(true); err != nil {
		return err
	}

	if _, err := self.Connection.Write(data); err != nil {
		return err
	}

	printCommand := cmdPrintAndEject
	if !cut {
		printCommand = cmdPrint
	}
	if _, err := self.Connection.Write(printCommand); err != nil {
		return err
	}

	return self.waitForPrintCompletion()
}

// reset sends the 100-byte invalidate preamble then initialize. NULL / ESC @ (reference.txt:767-791).
func (self *Printer) reset() error {
	if _, err := self.Connection.Write(make([]byte, 100)); err != nil {
		return err
	}
	_, err := self.Connection.Write(cmdInitialize)
	return err
}

// readStatus parses the 32-byte status response (reference.txt:813-855).
func (self *Printer) readStatus() (*Status, error) {
	buf := make([]byte, 32)
	if _, err := io.ReadFull(self.Connection, buf); err != nil {
		return nil, fmt.Errorf("read status: %w", err)
	}

	if len(buf) != 32 {
		return nil, fmt.Errorf("status must be 32 bytes, got: %d", len(buf))
	}

	return &Status{
		Type:       StatusType(buf[statusOffsetStatusType]),
		Model:      Model(buf[statusOffsetModel]),
		Battery:    BatteryStatusType(buf[statusOffsetBattery]),
		Error1:     Error1Type(buf[statusOffsetErrorInfo1]),
		Error2:     Error2Type(buf[statusOffsetErrorInfo2]),
		MediaType:  MediaType(buf[statusOffsetMediaType]),
		TapeColor:  TapeColor(buf[statusOffsetTapeColor]),
		TapeLength: int(buf[statusOffsetTapeLength]),
		TapeWidth:  TapeWidth(buf[statusOffsetMediaWidth]),
		FontColor:  FontColor(buf[statusOffsetFontColor]),
	}, nil
}

// waitForPrintCompletion polls the status type until printing ends. Table (5) (reference.txt:922-937).
func (self *Printer) waitForPrintCompletion() error {
	for {
		status, err := self.readStatus()
		if err != nil {
			return err
		}

		switch status.Type {
		case statusTypePrintingCompleted:
			return nil
		case statusTypeErrorOccurred:
			statusErrors := status.Errors()
			if len(statusErrors) == 0 {
				return fmt.Errorf("printer reported an unknown error")
			}
			return fmt.Errorf("printer reported: %s", strings.Join(statusErrors, ", "))
		case statusTypePowerOff:
			return fmt.Errorf("printer was turned off")
		default:
		}
	}
}

// setRasterMode switches the printer to raster mode. ESC i a (reference.txt:1032-1053).
func (self *Printer) setRasterMode() error {
	_, err := self.Connection.Write(cmdSetRasterMode)
	return err
}

// setNotificationMode toggles automatic status notifications. ESC i ! (reference.txt:1054-1073).
func (self *Printer) setNotificationMode(enabled bool) error {
	var value byte
	if enabled {
		value = 0x0
	} else {
		value = 0x1
	}

	payload := append(cmdNotifyModePrefix, value)

	_, err := self.Connection.Write(payload)
	return err
}

// setPrintProperty sets the print information. ESC i z (reference.txt:1074-1117).
func (self *Printer) setPrintProperty(rasterLines int, width TapeWidth) error {
	var enableFlag int

	enableFlag |= printPropertyEnableBitRecoverOnDevice

	tapeWidth := byte(width) //nolint:gosec // tape width is a single protocol byte (max 36mm)
	const tapeLength = byte(0x00)
	enableFlag |= printPropertyEnableBitWidth

	var rasterNum [4]byte
	binary.LittleEndian.PutUint32(rasterNum[:], uint32(rasterLines)) //nolint:gosec // line count fits the protocol's 4-byte field

	const mediaType = byte(0x00)
	const page = byte(0x00)
	const eeprom = byte(0x00)

	data := append(cmdSetPrintPropertyPrefix, []byte{
		byte(enableFlag), //nolint:gosec // composed of single-byte enable-bit constants
		mediaType,
		tapeWidth,
		tapeLength,
	}...)
	data = append(data, rasterNum[:]...)
	data = append(data, page, eeprom)

	_, err := self.Connection.Write(data)
	return err
}

// setPrintMode sets the various mode settings. ESC i M (reference.txt:1118-1133).
func (self *Printer) setPrintMode(autocut bool, mirror bool) error {
	var value int
	if autocut {
		value = util.SetBit(value, 6)
	}
	if mirror {
		value = util.SetBit(value, 7)
	}

	payload := append(cmdSetPrintModePrefix, byte(value))

	_, err := self.Connection.Write(payload)
	return err
}

// setExtendedMode sets the advanced mode settings. ESC i K (reference.txt:1134-1170).
func (self *Printer) setExtendedMode(noChainprint bool, specialTapeDisableCut bool, highDPI bool, noClearBuffer bool) error {
	var value int
	if noChainprint {
		value = util.SetBit(value, 3)
	}

	if specialTapeDisableCut {
		value = util.SetBit(value, 4)
	}

	if highDPI {
		value = util.SetBit(value, 6)
	}

	if noClearBuffer {
		value = util.SetBit(value, 7)
	}

	payload := append(cmdSetExtendedModePrefix, byte(value))

	_, err := self.Connection.Write(payload)
	return err
}

// setFeedAmount sets the margin/feed amount. ESC i d (reference.txt:1171-1203).
func (self *Printer) setFeedAmount(amount int) error {
	var bytes [2]byte
	binary.LittleEndian.PutUint16(bytes[:], uint16(amount)) //nolint:gosec // feed amount fits the protocol's 2-byte field

	payload := append(cmdSetFeedAmountPrefix, bytes[:]...)
	_, err := self.Connection.Write(payload)
	return err
}

// setCompressionModeEnabled selects the compression mode. M (reference.txt:1204-1277).
func (self *Printer) setCompressionModeEnabled(enabled bool) error {
	var value byte
	if enabled {
		value = 0x02
	}

	payload := append(cmdSetCompressionModePrefix, value)
	_, err := self.Connection.Write(payload)
	return err
}
