package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"runtime"

	_ "embed"

	"github.com/creack/pty"
	"github.com/lukasz-lobocki/termos/stage"
	"github.com/spf13/cobra"
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

/*
Cobra initiation.
*/
func init() {
	rootCmd.AddCommand(shotCmd)
}

func doShot(args []string) {
	var (
		err error
	)

	checkLogginglevel(args)

	s := stage.New()
	buf := getPrintout(TERMINAL_ROWS, TERMINAL_COLS, args[0], args[1:]...) // agr0 is the command to be run
	saveStream(buf.Bytes(), SAVED_STREAM_FILENAME)                         // save it // TODO make it sip from scaffold
	s.AddCommand(args...)                                                  // Add the issued command to the scaffold
	err = s.AddContent(&buf)                                               // Add the captured output to the scaffold
	check("failed adding content", err)
	if loggingLevel >= 3 {
		logInfo.Printf("from scaffold:\n%s", s.GetContent().String())
	}
	w, h := s.MeasureContent()
	logInfo.Printf("w: %f, h; %f", w, h)
	_, err = s.DoImage()
	check("imaging failed", err)

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
