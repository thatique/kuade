package commands

import (
	"github.com/spf13/cobra"

	"github.com/thatique/kuade/version"
)

// show application version number
var showVersion bool

// Execute execute our CLI application
func Execute() error {
	// Rootcmd is the rootcommand for
	var rootCmd = &cobra.Command{
		Use:   "Kuade",
		Short: "Kuade's CLI application to manage kuade application",
		Long:  `Thatique's CLI application to manage kuade applicaion.`,
		Run: func(cmd *cobra.Command, args []string) {
			if showVersion {
				version.PrintVersion()
				return
			}
			cmd.Usage()
		},
	}
	rootCmd.AddCommand(serveCommand())
	rootCmd.AddCommand(secretKeyCommand())

	return rootCmd.Execute()
}
