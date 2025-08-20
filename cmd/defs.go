package cmd

import (
	"image/color"

	"github.com/gonvenience/bunt"
	"golang.org/x/image/font"
)

const (
	TERMINAL_ROWS             = 40
	TERMINAL_COLS             = 120
	SAVED_STREAM_FILENAME     = "bigos.txt"
	MAX_LOGGING_LEVEL     int = 3 // Maximum allowed logging level.
	TERMINAL_COLOR            = "#151515"
	RED                       = "#ED655A"
	YELLOW                    = "#E1C04C"
	GREEN                     = "#71BD47"
)

type Scaffold struct {
	content                bunt.String
	factor                 float64
	columns                int
	padding                float64
	margin                 float64
	defaultForegroundColor color.Color
	tabSpaces              int

	regular    font.Face
	bold       font.Face
	italic     font.Face
	boldItalic font.Face

	lineSpacing float64
}
