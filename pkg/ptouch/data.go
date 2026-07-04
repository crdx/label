package ptouch

import "fmt"

// Raster Command Reference for PT-E550W/P750W/P710BT:
// https://download.brother.com/welcome/docp100064/cv_pte550wp750wp710bt_eng_raster_102.pdf

var (
	cmdInitialize = []byte{0x1b, 0x40}
	cmdDumpStatus = []byte{0x1b, 0x69, 0x53}
)

const (
	statusOffsetModel      = 4
	statusOffsetBattery    = 6
	statusOffsetMediaWidth = 10
	statusOffsetMediaType  = 11
	statusOffsetTapeLength = 17
	statusOffsetTapeColor  = 24
	statusOffsetFontColor  = 25
)

type Model int

const modelPTP710BT Model = 0x76 // PT-P710BT

type TapeWidth int

type MediaType int

const (
	mediaTypeNone         MediaType = 0    // No tape
	mediaTypeLaminated    MediaType = 0x01 // Laminated
	mediaTypeNonLaminated MediaType = 0x03 // Non laminated
	mediaTypeHeatShirink  MediaType = 0x11 // Heat shrink tube
	mediaTypeInvalid      MediaType = 0xFF // Invalid tape type
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
	case mediaTypeHeatShirink:
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

func unknown(value int) string {
	return fmt.Sprintf("Unknown (0x%02x)", value)
}
