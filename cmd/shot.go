package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/creack/pty"
	"github.com/fogleman/gg"
	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
)

// shotCmd represents the shell command.
var shotCmd = &cobra.Command{
	Short:   "Export ssh certificates.",
	Run:     func(cmd *cobra.Command, args []string) { shot(args) },
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

const (
	TERMINAL_ROWS         = 40
	TERMINAL_COLS         = 120
	SAVED_STREAM_FILENAME = "bigos.txt"
)

/*
Cobra initiation.
*/
func init() {
	rootCmd.AddCommand(shotCmd)
}

func shot(args []string) {
	buf := get_printout(TERMINAL_ROWS, TERMINAL_COLS, args[0], args[1:]...) // agr0 is the command to be run
	save_stream(buf.Bytes(), SAVED_STREAM_FILENAME)                         // save it
	graph(&buf)
}

func get_printout(rows uint16, cols uint16, cmd_name string, cmd_args ...string) bytes.Buffer {
	var b bytes.Buffer
	w := pty.Winsize{Rows: rows, Cols: cols}
	c := exec.Command(cmd_name, cmd_args...)
	f, err := pty.StartWithSize(c, &w) // get command output file from the (pty) pseudo-terminal
	check(err)
	io.Copy(&b, f) // read the stream, memorize it in the buffer
	return b
}

func graph(in io.Reader) {
	parsed, err := bunt.ParseStream(in)
	check(err)
	logInfo.Println(parsed.String())
	// for _, cr := range *parsed {
	// 	logInfo.Print(cr)
	// }
	dc := gg.NewContext(1000, 1000)
	dc.DrawCircle(500, 500, 400)
	dc.SetRGB(0, 0, 0)
	dc.Fill()
	err = dc.SavePNG("out.png")
	check(err)
}

func check(e error) {
	if e != nil {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			logError.Fatalf("File: %s Line:%v Message:%+v", file, line, e)
		} else {
			logError.Fatalf("Message:%+v", e)
		}
	}
}

func save_stream(source []byte, target_filename string) {
	o, err := os.Create(target_filename)
	check(err)
	defer o.Close()
	o.Write(source)
}
