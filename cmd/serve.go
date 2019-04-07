package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thatique/kuade/app"
	"github.com/thatique/kuade/app/handlers"
	"github.com/thatique/kuade/assets"
	"github.com/thatique/kuade/configuration"
)

var httpAddr string

var serveCommand = &cobra.Command{
	Use:   `serve`,
	Short: `Run kuade server`,
	Long: `Run Kuade's HTTP server. If you give it path to configuration.yml then it will
use it. Otherwise it try to load configuration from default path, system user or global path.`,
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := configuration.Get()
		if err != nil {
			fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
			cmd.Usage()
			os.Exit(1)
		}

		if httpAddr != "" {
			conf.HTTP.Addr = httpAddr
		}

		engine, err := app.NewEngine(handlers.NewApp(assets.Asset), conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating http server with error: %v", err)
			os.Exit(1)
		}
		if err = engine.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "application exit with error: %v", err)
			os.Exit(1)
		}
	},
}
