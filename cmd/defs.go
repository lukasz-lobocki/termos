package cmd

import (
	"github.com/gonvenience/bunt"
	"golang.org/x/image/font"
)

const (
	TERMINAL_ROWS             = 40
	TERMINAL_COLS             = 120
	SAVED_STREAM_FILENAME     = "bigos.txt"
	MAX_LOGGING_LEVEL     int = 3 // Maximum allowed logging level.
)

type Scaffold struct {
	content bunt.String
	factor  float64
	columns int
	padding float64
	margin  float64

	regular    font.Face
	bold       font.Face
	italic     font.Face
	boldItalic font.Face

	lineSpacing float64
}
