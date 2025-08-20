package stage

import (
	"fmt"
	"image"
	"image/color"
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
	TERMINAL_COLOR = "#151515"
	RED            = "#ED655A"
	YELLOW         = "#E1C04C"
	GREEN          = "#71BD47"
)

type Stage struct {
	content bunt.String
	factor  float64
	columns int
	padding float64
	margin  float64

	regular    font.Face
	bold       font.Face
	italic     font.Face
	boldItalic font.Face

	defaultForegroundColor color.Color
	backgroundColor        string
	lineSpacing            float64
	tabSpaces              int
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

func New() (Stage, error) { //TODO: male pierwsze litery klas?
	f := 1.0

	return Stage{
		factor:  f,
		margin:  0,      //f * 48, // empty area outside of terminal window // TODO make param
		padding: f * 24, // empty area inside of terminal window

		defaultForegroundColor: bunt.LightGray,
		backgroundColor:        TERMINAL_COLOR,

		columns: 120, // TODO parametrize

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

func (s *Stage) AddCommand(args ...string) {
	s.AddContent(strings.NewReader(
		bunt.Sprintf("Lime{$} Lime{%s}\n\n", strings.Join(args, " ")),
	))
}

func (s *Stage) MeasureContent() (width float64, height float64) {
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
		}
	default: // width based on column value
		width = float64(d.MeasureString(strings.Repeat("W", s.GetColumns())) >> 6) // W is the widest glyph
	}

	height = float64(len(ls)) * s.GetFontHeight() * s.lineSpacing

	return width, height // TODO w=w-1
}

func (s *Stage) DoImage() (image.Image, error) {
	var (
		f              = func(v float64) float64 { return s.factor * v }
		corner         = f(6)
		dotsRadius     = f(9)
		dotsDistance   = f(25)
		titleBarHeight = f(40) // TODO calculate instead of const
		marginX        = s.margin
		marginY        = s.margin
		paddingX       = s.padding
		paddingY       = s.padding
	)
	contentWidth, contentHeight := s.MeasureContent()
	contentWidth = math.Max(contentWidth, 3*dotsDistance+3*dotsRadius) // Make sure the output window is big enough

	width := contentWidth + 2*marginX + 2*paddingX
	height := contentHeight + 2*marginY + 2*paddingY + titleBarHeight
	dc := gg.NewContext(int(width), int(height))

	// Rounded rectangle inside the margins to produce an impression of a window
	dc.DrawRoundedRectangle(marginX, marginY, width-2*marginX, height-2*marginY, corner)
	dc.SetHexColor(s.backgroundColor)
	dc.Fill()

	// 3 colored dots mimicking menu bar
	for i, color := range []string{RED, YELLOW, GREEN} {
		dc.DrawCircle(marginX+paddingX+dotsRadius+float64(i)*dotsDistance, marginY+paddingY+dotsRadius, dotsRadius)
		dc.SetHexColor(color)
		dc.Fill()
	}

	// current posiion
	var x, y = marginX + paddingX, marginY + paddingY + titleBarHeight + s.GetFontHeight()

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
			dc.DrawRectangle(x, y-h+12, w, h)
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
			dc.SetColor(s.defaultForegroundColor)
		}

		// special symbols
		switch sym {
		case "\n":
			x = marginX + paddingX // reset x position
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

	err := dc.SavePNG("out.png") // TODO refactor, extract it outside
	if err != nil {
		return nil, fmt.Errorf("failed to save png. %w", err)
	}
	return dc.Image(), nil
}

func (s *Stage) WriteRaw(w io.Writer) error { // TODO use it
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
