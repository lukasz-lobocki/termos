package cmd

import (
	"bytes"
	"image"
	"io"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strings"

	_ "embed"

	"github.com/creack/pty"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/gonvenience/bunt"
	"github.com/gonvenience/term"
	"github.com/spf13/cobra"
	"golang.org/x/image/font"
)

// shotCmd represents the shell command.
var shotCmd = &cobra.Command{
	Short:   "Export ssh certificates.",
	Run:     func(cmd *cobra.Command, args []string) { doShot(args) },
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"take", "snag", "grab"},

	DisableFlagsInUseLine: true,

	Example: "  step-badger sshCerts ./db",
	Long: `
Export ssh certificates' data out of the badger database of step-ca.`,
	Use: `shot [shot flags] [--] command [command flags] [command arguments] [...] [flags]

Arguments:
  PATH   location of the source database`,
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

func NewScaffold() Scaffold { //TODO: wydziel jako osobny package i zmien na New, aby stowrzyc stage.New, male pierwsze litery klas.
	f := 1.0
	fontRegular, _ := truetype.Parse(MonoRegular)
	fontBold, _ := truetype.Parse(MonoBold)
	fontItalic, _ := truetype.Parse(MonoItalic)
	fontBoldItalic, _ := truetype.Parse(MonoBoldItalic)
	fontFaceOptions := &truetype.Options{Size: f * 12, DPI: 144}
	return Scaffold{
		factor:  f,
		margin:  0,      //f * 48, // empty area outside of terminal window // TODO make param
		padding: f * 24, // empty area inside of terminal window

		defaultForegroundColor: bunt.LightGray,
		regular:                truetype.NewFace(fontRegular, fontFaceOptions),
		bold:                   truetype.NewFace(fontBold, fontFaceOptions),
		italic:                 truetype.NewFace(fontItalic, fontFaceOptions),
		boldItalic:             truetype.NewFace(fontBoldItalic, fontFaceOptions),

		lineSpacing: 1.2,
		tabSpaces:   2,
	}
}

func (s *Scaffold) SetColumns(columns int) {
	s.columns = columns
}

func (s *Scaffold) GetColumns() (columns int) {
	if s.columns != 0 {
		return s.columns
	}
	columns, _ = term.GetTerminalSize()
	return columns
}

func (s *Scaffold) GetFontHeight() float64 {
	return float64(s.regular.Metrics().Height >> 6)
}

func (s *Scaffold) AddContent(in io.Reader) {
	var (
		bs bunt.String
		n  int // column counter
	)

	ps, err := bunt.ParseStream(in)
	check("failed to parse stream", err)

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
}

func (s *Scaffold) AddCommand(args ...string) {
	s.AddContent(strings.NewReader(
		bunt.Sprintf("Lime{$} Lime{%s}\n\n", strings.Join(args, " ")),
	))
}

func (s *Scaffold) measureContent() (width float64, height float64) {
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

func (s *Scaffold) DoImage() image.Image {
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
	contentWidth, contentHeight := s.measureContent()
	contentWidth = math.Max(contentWidth, 3*dotsDistance+3*dotsRadius) // Make sure the output window is big enough

	width := contentWidth + 2*marginX + 2*paddingX
	height := contentHeight + 2*marginY + 2*paddingY + titleBarHeight
	dc := gg.NewContext(int(width), int(height))

	// Rounded rectangle inside the margins to produce an impression of a window
	dc.DrawRoundedRectangle(marginX, marginY, width-2*marginX, height-2*marginY, corner)
	dc.SetHexColor(TERMINAL_COLOR)
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

		x += w // advance x position for the next symbol
	}

	err := dc.SavePNG("out.png")
	check("failed to save png", err)

	return dc.Image()
}

func (s *Scaffold) WriteRaw(w io.Writer) { // TODO use it
	_, err := w.Write([]byte(s.content.String()))
	check("writing raw failed", err)
}

/*
Cobra initiation.
*/
func init() {
	rootCmd.AddCommand(shotCmd)
}

func doShot(args []string) {

	checkLogginglevel(args)

	s := NewScaffold()
	buf := getPrintout(TERMINAL_ROWS, TERMINAL_COLS, args[0], args[1:]...) // agr0 is the command to be run
	saveStream(buf.Bytes(), SAVED_STREAM_FILENAME)                         // save it // TODO make it sip from scaffold
	s.AddCommand(args...)                                                  // Add the issued command to the scaffold
	s.AddContent(&buf)                                                     // Add the captured output to the scaffold
	if loggingLevel >= 3 {
		logInfo.Printf("from scaffold:\n%s", s.content.String())
	}
	w, h := s.measureContent()
	logInfo.Printf("w: %f, h; %f", w, h)
	s.DoImage()
}

func getPrintout(rows uint16, cols uint16, cmd_name string, cmd_args ...string) (printout bytes.Buffer) {
	w := pty.Winsize{Rows: rows, Cols: cols} // size using received parameters
	c := exec.Command(cmd_name, cmd_args...)
	f, err := pty.StartWithSize(c, &w) // get command output file from the (pty) pseudo-terminal
	check("failed to read pseudo-terminal file", err)
	io.Copy(&printout, f) // read the stream, memorize it in the buffer
	return printout
}

func check(hint string, e error) {
	if len(hint) == 0 {
		hint = "no hint"
	}
	if e != nil {
		_, f, l, ok := runtime.Caller(1)
		if ok {
			logError.Fatalf("%s; %+v @ %s # %d", hint, e, f, l)
		} else {
			logError.Fatalf("%s; %+v", hint, e)
		}
	}
}

func saveStream(source []byte, target_filename string) {
	o, err := os.Create(target_filename)
	check("failed to create a file", err)
	defer o.Close()
	o.Write(source)
}
