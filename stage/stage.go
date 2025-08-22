package stage

import (
	"fmt"
	"image"
	"io"
	"math"
	"strings"

	_ "embed"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/gonvenience/bunt"
	"github.com/gonvenience/term"
	"golang.org/x/image/font"
)

const (
	RED    = "#ED655A"
	YELLOW = "#E1C04C"
	GREEN  = "#71BD47"
)

type Stage struct {
	content bunt.String
	factor  float64
	columns int
	padding float64

	regular    font.Face
	bold       font.Face
	italic     font.Face
	boldItalic font.Face

	titlebarColor   string
	backgroundColor string
	foregroundColor string
	commandColor    string
	lineSpacing     float64
	tabSpaces       int
}

var (
	//go:embed JuliaMono-Bold.ttf
	MonoBold []byte
	//go:embed JuliaMono-BoldItalic.ttf
	MonoBoldItalic []byte
	//go:embed JuliaMono-RegularItalic.ttf
	MonoItalic []byte
	//go:embed JuliaMono-Regular.ttf
	MonoRegular []byte
)

func New(titlebarColor string, backgroundColor string, foregroundColor string, commandColor string, magnification int, cols int) (Stage, error) {
	f := float64(magnification)

	return Stage{
		factor:  f,
		padding: f * 24, // empty area inside of terminal window

		titlebarColor:   titlebarColor,
		backgroundColor: backgroundColor,
		foregroundColor: foregroundColor,
		commandColor:    commandColor,

		columns: cols,

		lineSpacing: 1.2,
		tabSpaces:   2,
	}, nil
}

func (s *Stage) AddFonts() error {
	fontFaceOptions := &truetype.Options{Size: s.factor * 12, DPI: 144}

	fontRegular, err := truetype.Parse(MonoRegular)
	if err != nil {
		return fmt.Errorf("failed to parse MonoRegular font. %w", err)
	}
	s.regular = truetype.NewFace(fontRegular, fontFaceOptions)

	fontBold, err := truetype.Parse(MonoBold)
	if err != nil {
		return fmt.Errorf("failed to parse MonoBold font. %w", err)
	}
	s.bold = truetype.NewFace(fontBold, fontFaceOptions)

	fontItalic, err := truetype.Parse(MonoItalic)
	if err != nil {
		return fmt.Errorf("failed to parse MonoItalic font. %w", err)
	}
	s.italic = truetype.NewFace(fontItalic, fontFaceOptions)

	fontBoldItalic, err := truetype.Parse(MonoBoldItalic)
	if err != nil {
		return fmt.Errorf("failed to parse MonoBoldItalic font. %w", err)
	}
	s.boldItalic = truetype.NewFace(fontBoldItalic, fontFaceOptions)

	return nil
}

func (s *Stage) AddContent(in io.Reader) error {
	var (
		bs bunt.String
		n  int // column counter
	)

	ps, err := bunt.ParseStream(in)
	if err != nil {
		return fmt.Errorf("failed to parse stream. %w", err)
	}

	for _, cr := range *ps { // wrap if wider than capcity

		n++

		if cr.Symbol == '\n' { // reset when new line encountered
			n = 0
		}
		if n > s.GetColumns() { // wrap in case the column capacity is reached and reset the counter
			bs = append(bs, bunt.ColoredRune{Settings: cr.Settings, Symbol: '\n'})
			n = 0
		}

		bs = append(bs, cr)
	}
	s.content = append(s.content, bs...)
	return nil
}

func (s *Stage) AddCommand(args ...string) error {
	err := s.AddContent(strings.NewReader(
		bunt.Sprintf(s.commandColor+"{$} "+s.commandColor+"{%s}\n\n", strings.Join(args, " ")),
	))
	return err
}

func (s *Stage) MeasureContent() (width float64, height float64, columns int) {
	var (
		rc = make([]rune, len(s.content))
	)
	for i, cr := range s.content { // extract symbols from content
		rc[i] = cr.Symbol
	}

	ls := strings.Split(
		strings.TrimSuffix(string(rc), "\n"), // avoid unnecessary split at the very end
		"\n",                                 // by lines
	)

	// temporary drawer for measurements
	d := &font.Drawer{Face: s.regular}

	switch s.columns {
	case 0: // width based on actual longest line
		for _, l := range ls {
			if lw := float64(d.MeasureString(l) >> 6); lw > width { // type of fixed.Int26_6 divided by 2^6
				width = lw // update width if measured current line width was bigger
			}
			if cw := bunt.PlainTextLength(l); cw > columns {
				columns = cw // update columns if measured current line columns was bigger
			}
		}
	default: // width based on column value
		width = float64(d.MeasureString(strings.Repeat("W", s.GetColumns())) >> 6) // W is the widest glyph
		columns = s.GetColumns()
	}

	height = float64(len(ls)) * s.GetFontHeight() * s.lineSpacing

	return width, height, columns
}

func (s *Stage) GetImage(contentWidth float64, contentHeight float64) image.Image {
	var (
		f              = func(v float64) float64 { return s.factor * v }
		paddingX       = s.padding
		paddingY       = s.padding - f(10)
		corner         = f(6)
		dotsRadius     = f(9)
		dotsDistance   = dotsRadius * 3
		titleBarHeight = dotsRadius*2 + paddingY*2
	)
	contentWidth = math.Max(contentWidth, 3*dotsDistance+3*dotsRadius) // make sure the output window is big enough

	width := contentWidth + 2*paddingX
	height := contentHeight + 2*paddingY + titleBarHeight
	dc := gg.NewContext(int(width), int(height))

	// Rounded rectangle to produce an impression of a window
	dc.DrawRoundedRectangle(0, corner, width, height-corner, corner) // lowered by the corner to hide bacground artifacts from behind
	dc.SetHexColor(s.backgroundColor)
	dc.Fill()

	// Semi rectangle to produce an impression of a titlebar
	dc.DrawRoundedRectangle(0, 0, width, titleBarHeight, corner)
	dc.SetHexColor(s.titlebarColor)
	dc.Fill()
	dc.DrawRectangle(0, corner, width, titleBarHeight-corner) // making bottom flat
	dc.SetHexColor(s.titlebarColor)
	dc.Fill()

	// 3 colored dots mimicking menu bar
	for i, color := range []string{RED, YELLOW, GREEN} {
		dc.DrawCircle(paddingX+dotsRadius+float64(i)*dotsDistance, titleBarHeight/2, dotsRadius)
		dc.SetHexColor(color)
		dc.Fill()
	}

	// current posiion
	var x, y = paddingX, paddingY + titleBarHeight + s.GetFontHeight()

	for _, cr := range s.content { // for each rune

		// change font face
		switch cr.Settings & 0x1C {
		case 4:
			dc.SetFontFace(s.bold)
		case 8:
			dc.SetFontFace(s.italic)
		case 12:
			dc.SetFontFace(s.boldItalic)
		default:
			dc.SetFontFace(s.regular)
		}

		sym := string(cr.Symbol)
		w, h := dc.MeasureString(sym)

		// change background color
		switch cr.Settings & 0x02 {
		case 2:
			dc.SetRGB255(
				int((cr.Settings>>32)&0xFF),
				int((cr.Settings>>40)&0xFF),
				int((cr.Settings>>48)&0xFF),
			)
			dc.DrawRectangle(x, y-h+f(3), w, h+f(3)) // f(3) added to account for ascenders & descenders
			dc.Fill()
		}

		// change foreground color
		switch cr.Settings & 0x01 {
		case 1:
			dc.SetRGB255(
				int((cr.Settings>>8)&0xFF),
				int((cr.Settings>>16)&0xFF),
				int((cr.Settings>>24)&0xFF),
			)
		default:
			dc.SetHexColor(s.foregroundColor)
		}

		// special symbols
		switch sym {
		case "\n":
			x = paddingX           // reset x position
			y += h * s.lineSpacing // advance y position by line spacing
			continue
		case "\t":
			x += w * float64(s.tabSpaces) // advance x position by tab
			continue
		case "✗", "ˣ": // mitigate issue #1 by replacing it with a similar character
			sym = "×"
		}

		dc.DrawString(sym, x, y)

		// manually draw an underline under each character
		if cr.Settings&0x1C == 16 {
			dc.DrawLine(x, y+f(4), x+w, y+f(4))
			dc.SetLineWidth(f(1))
			dc.Stroke()
		}

		x += w // advance the x position for the next symbol
	}
	return dc.Image()
}

func (s *Stage) WriteRaw(w io.Writer) error {
	_, err := w.Write([]byte(s.content.String()))
	if err != nil {
		return fmt.Errorf("writing raw failed. %w", err)
	}
	return nil
}

func (s *Stage) SetColumns(columns int) {
	s.columns = columns
}

func (s *Stage) GetColumns() (columns int) {
	if s.columns != 0 {
		return s.columns
	}
	columns, _ = term.GetTerminalSize()
	return columns
}

func (s *Stage) GetFontHeight() float64 {
	return float64(s.regular.Metrics().Height >> 6)
}

func (s *Stage) GetContent() bunt.String {
	return s.content
}
