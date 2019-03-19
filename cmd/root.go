package cmd

import (
	"github.com/spf13/cobra"

	"github.com/thatique/kuade/version"
)

var showVersion bool

var RootCmd = &cobra.Command{
	Use:   "Thatique",
	Short: "Thatique's CLI application to manage thatique server",
	Long:  `Thatique's CLI application to manage thatique server.`,
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			version.PrintVersion()
			return
		}
		cmd.Usage()
	},
}
