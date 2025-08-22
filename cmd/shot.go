package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	_ "embed"

	"github.com/creack/pty"
	"github.com/fogleman/gg"
	"github.com/gonvenience/term"
	"github.com/lukasz-lobocki/termos/stage"
	"github.com/spf13/cobra"
)

// shotCmd represents the shell command.
var shotCmd = &cobra.Command{
	Short:   "Create a screenshot",
	Run:     func(cmd *cobra.Command, args []string) { doShot(args) },
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"take", "snag", "grab"},

	DisableFlagsInUseLine: true,

	Example: "  termos shot --columns 80 -- printf '1234567890%.0s' {1..6}",
	Long: `
Create png and txt color screenshots of the terminal command output.`,
	Use: `shot [shot flags] [--] command [command flags] [command arguments] [...] [flags]`,
}

/*
Cobra initiation.
*/
func init() {
	// Hide help command.
	shotCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	//Do not sort flags.
	shotCmd.Flags().SortFlags = false

	shotCmd.Flags().IntVarP(&config.columnNumber, "columns", "c", 0, "number of columns rendered (default auto)")
	shotCmd.Flags().StringVarP(&config.savingFilename, "filename", "f", "out", "name of files to be saved")
	shotCmd.Flags().IntVarP(&config.magnification, "magnification", "m", 1, "magnification factor")
	shotCmd.Flags().StringVar(&config.titlebarColor, "tc", "#696969", "titlebar color hex")
	shotCmd.Flags().StringVar(&config.backgroundColor, "bc", "#151515", "background color hex")
	shotCmd.Flags().StringVar(&config.foregroundColor, "fc", "#DCDCDC", "foreground color hex")
	shotCmd.Flags().StringVar(&config.commandColor, "cc", "#16AF1B", "command color hex")
}

func doShot(args []string) {
	var (
		err error
		s   stage.Stage
	)
	checkLogginglevel(args)

	terminalWidth, terminalHeight := term.GetTerminalSize()
	if loggingLevel >= 1 {
		logInfo.Printf("pseudo-terminal width=%d, height=%d", terminalWidth, terminalHeight)
	}
	buf, err := getTerminalOutput(terminalHeight, terminalWidth, args[0], args[1:]...) // agr0 is the command to be run
	if err != nil {
		logError.Fatalf("failed getting printout. %v", err)
	}

	s, err = stage.New(config.titlebarColor, config.backgroundColor, config.foregroundColor, config.commandColor, config.magnification, config.columnNumber)
	if err != nil {
		logError.Fatalf("failed creating stage. %v+", err)
	}
	err = s.AddFonts()
	if err != nil {
		logError.Fatalf("failed adding fonts. %v+", err)
	}

	err = s.AddCommand(args...) // Add the issued command to the scaffold
	if err != nil {
		logError.Fatalf("failed adding command. %v+", err)
	}
	err = s.AddContent(&buf) // Add the captured output to the scaffold
	if err != nil {
		logError.Fatalf("failed adding content. %v+", err)
	}

	contentWidth, contentHeight, contentColumns := s.MeasureContent()
	logInfo.Printf("Number of columns used: %d. Use '--columns' flag to impose it.", contentColumns)
	img := s.GetImage(contentWidth, contentHeight)
	if err != nil {
		logError.Fatalf("imaging failed. %v+", err)
	}

	err = saveStage(filepath.Clean(config.savingFilename+".txt"), s)
	if err != nil {
		logError.Fatalf("failed saving stage. %v+", err)
	}

	err = gg.SavePNG(filepath.Clean(config.savingFilename+".png"), img)
	if err != nil {
		logError.Fatalf("failed saving png. %v+", err)
	}

}

func saveStage(path string, s stage.Stage) error {
	if loggingLevel >= 2 {
		logInfo.Printf("Saving content to %s", path)
	}
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = output.Close() }()
	err = s.WriteRaw(output)
	return err
}

func getTerminalOutput(rows int, cols int, cmd_name string, cmd_args ...string) (printout bytes.Buffer, err error) {
	w := pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)} // size using received parameters
	c := exec.Command(cmd_name, cmd_args...)
	f, err := pty.StartWithSize(c, &w) // get command output file from the (pty) pseudo-terminal
	if err != nil {
		return bytes.Buffer{}, err
	}
	io.Copy(&printout, f) // read the stream, memorize it in the buffer
	return printout, nil
}
