package ptouch

import "fmt"

// Raster Command Reference for PT-E550W/P750W/P710BT:
// https://download.brother.com/welcome/docp100064/cv_pte550wp750wp710bt_eng_raster_102.pdf

var (
	cmdInitialize               = []byte{0x1b, 0x40}
	cmdDumpStatus               = []byte{0x1b, 0x69, 0x53}
	cmdSetRasterMode            = []byte{0x1b, 0x69, 0x61, 0x01} // 0: ESC/P, 1: Raster, 3: P-touch Template
	cmdNotifyModePrefix         = []byte{0x1b, 0x69, 0x21}
	cmdSetPrintPropertyPrefix   = []byte{0x1b, 0x69, 0x7a}
	cmdSetPrintModePrefix       = []byte{0x1b, 0x69, 0x4d}
	cmdSetExtendedModePrefix    = []byte{0x1b, 0x69, 0x4b}
	cmdSetFeedAmountPrefix      = []byte{0x1b, 0x69, 0x64}
	cmdSetCompressionModePrefix = []byte{0x4d}
	cmdRasterTransfer           = []byte{0x47}
	cmdPrint                    = []byte{0x0c} // FF: print without feeding (non-last page)
	cmdPrintAndEject            = []byte{0x1a} // Control-Z: print with feeding (last page)
)

const (
	statusOffsetModel      = 4
	statusOffsetBattery    = 6
	statusOffsetErrorInfo1 = 8
	statusOffsetErrorInfo2 = 9
	statusOffsetMediaWidth = 10
	statusOffsetMediaType  = 11
	statusOffsetTapeLength = 17
	statusOffsetStatusType = 18
	statusOffsetTapeColor  = 24
	statusOffsetFontColor  = 25
)

const (
	printPropertyEnableBitWidth           = 0x04
	printPropertyEnableBitRecoverOnDevice = 0x80
)

type Model int

const modelPTP710BT Model = 0x76 // PT-P710BT

type StatusType int

const (
	statusTypePrintingCompleted StatusType = 0x01 // Printing completed
	statusTypeErrorOccurred     StatusType = 0x02 // Error occurred
	statusTypePowerOff          StatusType = 0x04 // Power off
)

type Error1Type int

const (
	error1NoMedia          Error1Type = 0x01 // No media
	error1CutterJam        Error1Type = 0x04 // Cutter jam
	error1WeakBattery      Error1Type = 0x08 // Weak battery
	error1TooHighVoltageAC Error1Type = 0x40 // High-voltage adapter
)

type Error2Type int

const (
	error2InvalidMedia Error2Type = 0x01 // Invalid media
	error2CoverOpen    Error2Type = 0x10 // Cover open
	error2Hot          Error2Type = 0x20 // Too hot
)

type TapeWidth int

type MediaType int

const (
	mediaTypeNone           MediaType = 0    // No tape
	mediaTypeLaminated      MediaType = 0x01 // Laminated
	mediaTypeNonLaminated   MediaType = 0x03 // Non-laminated
	mediaTypeHeatShrink2To1 MediaType = 0x11 // Heat shrink tube (HS 2:1)
	mediaTypeHeatShrink3To1 MediaType = 0x17 // Heat shrink tube (HS 3:1)
	mediaTypeInvalid        MediaType = 0xFF // Invalid tape type
)

type TapeColor int

const (
	tapeColorWhite             TapeColor = 0x01
	tapeColorOther             TapeColor = 0x02
	tapeColorClear             TapeColor = 0x03
	tapeColorRed               TapeColor = 0x04
	tapeColorBlue              TapeColor = 0x05
	tapeColorYellow            TapeColor = 0x06
	tapeColorGreen             TapeColor = 0x07
	tapeColorBlack             TapeColor = 0x08
	tapeColorClearWhiteText    TapeColor = 0x09
	tapeColorMatteWhite        TapeColor = 0x20
	tapeColorMatteClear        TapeColor = 0x21
	tapeColorMatteSilver       TapeColor = 0x22
	tapeColorSatinGold         TapeColor = 0x23
	tapeColorSatinSilver       TapeColor = 0x24
	tapeColorDBlue             TapeColor = 0x30 // TZe-535, TZe-545, TZe-555
	tapeColorDRed              TapeColor = 0x31 // TZe-435
	tapeColorFluorescentOrange TapeColor = 0x40
	tapeColorFluorescentyellow TapeColor = 0x41
	tapeColorBerryPink         TapeColor = 0x50 // TZe-MQP35
	tapeColorLightGray         TapeColor = 0x51 // TZe-MQL35
	tapeColorLimeGreen         TapeColor = 0x52 // TZe-MQG35
	tapeColorFYellow           TapeColor = 0x60
	tapeColorFPing             TapeColor = 0x61
	tapeColorFBlue             TapeColor = 0x62
	tapeColorHeatShrinkWhite   TapeColor = 0x70
	tapeColorFlexWhite         TapeColor = 0x90
	tapeColorFlexYellow        TapeColor = 0x91
	tapeColorCleaning          TapeColor = 0xF0
	tapeColorStencil           TapeColor = 0xF1
	tapeColorInvalid           TapeColor = 0xFF
)

type FontColor int

const (
	fontColorWhite    FontColor = 0x01
	fontColorRed      FontColor = 0x04
	fontColorBlue     FontColor = 0x05
	fontColorBlack    FontColor = 0x08
	fontColorGold     FontColor = 0x0a
	fontColorFBlue    FontColor = 0x62
	fontColorCleaning FontColor = 0xF0
	fontColorStencil  FontColor = 0xF1
	fontColorOther    FontColor = 0x02
	fontColorInvalid  FontColor = 0xFF
)

type BatteryStatusType int

const (
	batteryFull            BatteryStatusType = 0
	batteryHalf            BatteryStatusType = 1
	batteryLow             BatteryStatusType = 2
	batteryChangeBatteries BatteryStatusType = 3
	batteryAC              BatteryStatusType = 4
)

func (self Model) String() string {
	if self == modelPTP710BT {
		return "PT-P710BT"
	}
	return unknown(int(self))
}

func (self BatteryStatusType) String() string {
	switch self {
	case batteryFull:
		return "Full"
	case batteryHalf:
		return "Half"
	case batteryLow:
		return "Low"
	case batteryChangeBatteries:
		return "Critical"
	case batteryAC:
		return "AC"
	default:
		return unknown(int(self))
	}
}

func (self MediaType) String() string {
	switch self {
	case mediaTypeNone:
		return "No tape"
	case mediaTypeLaminated:
		return "Laminated"
	case mediaTypeNonLaminated:
		return "Non-laminated"
	case mediaTypeHeatShrink2To1, mediaTypeHeatShrink3To1:
		return "Heat shrink tube"
	case mediaTypeInvalid:
		return "Invalid"
	default:
		return unknown(int(self))
	}
}

func (self TapeColor) String() string {
	switch self {
	case tapeColorWhite:
		return "White"
	case tapeColorOther:
		return "Other"
	case tapeColorClear:
		return "Clear"
	case tapeColorRed:
		return "Red"
	case tapeColorBlue:
		return "Blue"
	case tapeColorYellow:
		return "Yellow"
	case tapeColorGreen:
		return "Green"
	case tapeColorBlack:
		return "Black"
	case tapeColorClearWhiteText:
		return "Clear (white text)"
	case tapeColorMatteWhite:
		return "Matte white"
	case tapeColorMatteClear:
		return "Matte clear"
	case tapeColorMatteSilver:
		return "Matte silver"
	case tapeColorSatinGold:
		return "Satin gold"
	case tapeColorSatinSilver:
		return "Satin silver"
	case tapeColorDBlue:
		return "Dark blue"
	case tapeColorDRed:
		return "Dark red"
	case tapeColorFluorescentOrange:
		return "Fluorescent orange"
	case tapeColorFluorescentyellow:
		return "Fluorescent yellow"
	case tapeColorBerryPink:
		return "Berry pink"
	case tapeColorLightGray:
		return "Light grey"
	case tapeColorLimeGreen:
		return "Lime green"
	case tapeColorFYellow:
		return "Flexible yellow"
	case tapeColorFPing:
		return "Flexible pink"
	case tapeColorFBlue:
		return "Flexible blue"
	case tapeColorHeatShrinkWhite:
		return "Heat shrink white"
	case tapeColorFlexWhite:
		return "Flexible white"
	case tapeColorFlexYellow:
		return "Flexible yellow"
	case tapeColorCleaning:
		return "Cleaning"
	case tapeColorStencil:
		return "Stencil"
	case tapeColorInvalid:
		return "Invalid"
	default:
		return unknown(int(self))
	}
}

func (self FontColor) String() string {
	switch self {
	case fontColorWhite:
		return "White"
	case fontColorRed:
		return "Red"
	case fontColorBlue:
		return "Blue"
	case fontColorBlack:
		return "Black"
	case fontColorGold:
		return "Gold"
	case fontColorFBlue:
		return "Flexible blue"
	case fontColorCleaning:
		return "Cleaning"
	case fontColorStencil:
		return "Stencil"
	case fontColorOther:
		return "Other"
	case fontColorInvalid:
		return "Invalid"
	default:
		return unknown(int(self))
	}
}

func (self *Status) Errors() []string {
	var errors []string

	if self.Error1&error1NoMedia != 0 {
		errors = append(errors, "No media")
	}
	if self.Error1&error1CutterJam != 0 {
		errors = append(errors, "Cutter jam")
	}
	if self.Error1&error1WeakBattery != 0 {
		errors = append(errors, "Weak battery")
	}
	if self.Error1&error1TooHighVoltageAC != 0 {
		errors = append(errors, "AC voltage too high")
	}

	if self.Error2&error2InvalidMedia != 0 {
		errors = append(errors, "Invalid media")
	}
	if self.Error2&error2CoverOpen != 0 {
		errors = append(errors, "Cover open")
	}
	if self.Error2&error2Hot != 0 {
		errors = append(errors, "Too hot")
	}

	return errors
}

func (self *Status) PrintableDots() (int, error) {
	switch self.MediaType {
	case mediaTypeLaminated, mediaTypeNonLaminated:
	case mediaTypeNone:
		return 0, fmt.Errorf("no tape loaded")
	case mediaTypeHeatShrink2To1, mediaTypeHeatShrink3To1:
		return 0, fmt.Errorf("heat shrink tube is not supported")
	default:
		return 0, fmt.Errorf("unsupported media type: %s", self.MediaType)
	}

	switch self.TapeWidth {
	case 4:
		return 24, nil
	case 6:
		return 32, nil
	case 9:
		return 50, nil
	case 12:
		return 70, nil
	case 18:
		return 112, nil
	case 24:
		return 128, nil
	default:
		return 0, fmt.Errorf("unsupported tape width: %dmm", self.TapeWidth)
	}
}

func unknown(value int) string {
	return fmt.Sprintf("Unknown (0x%02x)", value)
}
