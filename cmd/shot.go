package cmd

import (
	"bytes"
	"image"
	"io"
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

func NewScaffold() Scaffold { //TODO: wydziel jako osobny package i zmien na New, aby stowrzyc package.new
	f := 1.0
	fontRegular, _ := truetype.Parse(MonoRegular)
	fontBold, _ := truetype.Parse(MonoBold)
	fontItalic, _ := truetype.Parse(MonoItalic)
	fontBoldItalic, _ := truetype.Parse(MonoBoldItalic)
	fontFaceOptions := &truetype.Options{Size: f * 12, DPI: 144}
	return Scaffold{
		factor:      f,
		margin:      0, //f * 48,
		padding:     f * 24,
		regular:     truetype.NewFace(fontRegular, fontFaceOptions),
		bold:        truetype.NewFace(fontBold, fontFaceOptions),
		italic:      truetype.NewFace(fontItalic, fontFaceOptions),
		boldItalic:  truetype.NewFace(fontBoldItalic, fontFaceOptions),
		lineSpacing: 1.2,
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

func (s *Scaffold) FontHeight() float64 {
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

	height = float64(len(ls)) * s.FontHeight() * s.lineSpacing

	return width, height
}

func (s *Scaffold) image() image.Image {
	// var (
	// 	f              = func(v float64) float64 { return s.factor * v }
	// 	corner         = f(6)
	// 	radius         = f(9)
	// 	distance       = f(25) // TODO: get rid
	// 	titleBarHeight = f(40)
	// )
	// contentWidth, contentHeight := s.measureContent()
	// contentWidth = math.Max(contentWidth, 3*distance+3*radius) // Make sure the output window is big enough

	// marginX, marginY := s.margin, s.margin
	// xOffset, yOffset := marginX, marginY
	// paddingX, paddingY := s.padding, s.padding

	// width := contentWidth + 2*marginX + 2*paddingX
	// height := contentHeight + 2*marginY + 2*paddingY + titleBarHeight
	dc := gg.NewContext(2, 2) // dc := gg.NewContext(int(width), int(height))

	// Draw rounded rectangle with outline to produce impression of a window
	//
	// dc.DrawRoundedRectangle(xOffset, yOffset, width-2*marginX, height-2*marginY, corner)
	// dc.SetHexColor("#151515")
	// dc.Fill()

	return dc.Image()
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
	s.AddContent(&buf)                                                     // Add the captured output to the scaffold
	if loggingLevel >= 3 {
		logInfo.Printf("from scaffold:\n%s", s.content.String())
	}
	//graph(&buf)
}

func getPrintout(rows uint16, cols uint16, cmd_name string, cmd_args ...string) (printout bytes.Buffer) {
	w := pty.Winsize{Rows: rows, Cols: cols}
	c := exec.Command(cmd_name, cmd_args...)
	f, err := pty.StartWithSize(c, &w) // get command output file from the (pty) pseudo-terminal
	check("failed to read pseudo-terminal file", err)
	io.Copy(&printout, f) // read the stream, memorize it in the buffer
	return printout
}

func graph(in io.Reader) {
	parsed, err := bunt.ParseStream(in)
	check("failed to parse stream", err)
	if loggingLevel >= 3 {
		logInfo.Println(parsed.String())
	}

	dc := gg.NewContext(1000, 1000)
	dc.DrawCircle(500, 500, 400)
	dc.SetRGB(0, 0, 0)
	dc.Fill()
	err = dc.SavePNG("out.png")
	check("failed to save png", err)
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
