package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/fogleman/gg"
	"github.com/lukasz-lobocki/termos/stage"
	"github.com/spf13/cobra"
)

// shotCmd represents the shell command.
var renderCmd = &cobra.Command{
	Short:   "Render a screenshot",
	Run:     func(cmd *cobra.Command, args []string) { doRender(args) },
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"draw"},

	DisableFlagsInUseLine: true,

	Example: "  termos render --columns 80 -- out.txt",
	Long: `
Render png color screenshot of the file input.`,
	Use: `render [render flags] [--] filename`,
}

/*
Cobra initiation.
*/
func init() {
	// Hide help command.
	renderCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	//Do not sort flags.
	renderCmd.Flags().SortFlags = false

	renderCmd.Flags().IntVarP(&config.columnNumber, "columns", "c", 0, "number of columns rendered (default auto)")
	renderCmd.Flags().IntVarP(&config.magnification, "magnification", "m", 1, "magnification factor")
	renderCmd.Flags().StringVar(&config.titlebarColor, "tc", "#696969", "titlebar color hex")
	renderCmd.Flags().StringVar(&config.backgroundColor, "bc", "#151515", "background color hex")
	renderCmd.Flags().StringVar(&config.foregroundColor, "fc", "#DCDCDC", "foreground color hex")
	renderCmd.Flags().StringVar(&config.commandColor, "cc", "#16AF1B", "command color hex")
}

func doRender(args []string) {
	var (
		err error
		s   stage.Stage
	)
	checkLogginglevel(args)

	buf, err := getFileOutput(filepath.Clean(args[0]))
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

	err = gg.SavePNG(filepath.Clean(strings.TrimSuffix(args[0], filepath.Ext(args[0]))+".png"), img)
	if err != nil {
		logError.Fatalf("failed saving png. %v+", err)
	}

}

func getFileOutput(filename string) (printout bytes.Buffer, err error) {
	if loggingLevel >= 2 {
		logInfo.Printf("Getting content from %s", filename)
	}
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return printout, err
	}
	_, err = printout.Write(bytes)
	return printout, err
}
