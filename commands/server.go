package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/thatique/kuade/app/config"
	"github.com/thatique/kuade/app/handlers"
	"github.com/thatique/kuade/app/storage"
)

func serveCommand() *cobra.Command {
	vsrv := viper.New()
	cfg := handlers.DefaultAppConfig()

	serveCommand := &cobra.Command{
		Use:   "serve",
		Short: "start our application server",
		Long:  "serve is used to start kuade application server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// initialize
			storage.DefaultURLMux().InitFromViper(vsrv)
			cfg.InitFromViper(vsrv)

			return nil
		},
	}

	config.AddFlags(vsrv, serveCommand, cfg.AddFlags, storage.DefaultURLMux().AddFlags)

	return serveCommand
}
