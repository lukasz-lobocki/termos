package cmd

import "github.com/spf13/cobra"

// buildinfoCmd represents the shell command.
var buildinfoCmd = &cobra.Command{
	Short: "Export ssh certificates.",
	Run:   func(cmd *cobra.Command, args []string) { getbuildinfo() },
	Args:  cobra.NoArgs,

	DisableFlagsInUseLine: true,

	Example: "  step-badger sshCerts ./db",
	Long: `
Export ssh certificates' data out of the badger database of step-ca.`,
	Use: `buildinfo <PATH> [flags]

Arguments:
  PATH   location of the source database`,
}

/*
Cobra initiation.
*/
func init() {
	rootCmd.AddCommand(buildinfoCmd)
}

func getbuildinfo() {
	println("release version:", semReleaseVersion)
	println()
	println("semVer:", semVer)
	println("commitHash:", commitHash)
	println("isGitDirty:", isGitDirty)
	println("isSnapshot:", isSnapshot)
	println("goOs:", goOs)
	println("goArch:", goArch)
	println("gitUrl:", gitUrl)
	println("builtBranch:", builtBranch)
	println("builtDate:", builtDate)
}
