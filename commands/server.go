package commands

import (
	"context"
	"flag"
	"fmt"
	"os"

	"contrib.go.opencensus.io/exporter/ocagent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gocloud.dev/requestlog"
	"gocloud.dev/server"

	"github.com/thatique/kuade/app"
	"github.com/thatique/kuade/app/config"
	"github.com/thatique/kuade/app/handlers"
	"github.com/thatique/kuade/app/storage"
	"github.com/thatique/kuade/assets"
	scontext "github.com/thatique/kuade/pkg/web/context"
	"github.com/thatique/kuade/version"
)

const (
	webPort     = "port"
	storageURL  = "storage-url"
	ocAgentAddr = "ocagent-addr"
)

type serverOption struct {
	storageURL  string
	addr        string
	ocAgentAddr string
}

func (opt *serverOption) AddFlags(flagSet *flag.FlagSet) {
	flagSet.String(
		webPort,
		opt.addr,
		"The addr and port to use to serve")
	flagSet.String(
		storageURL,
		opt.storageURL,
		"the storage driver URL you want to open")
	flagSet.String(
		"ocagent-addr",
		opt.ocAgentAddr,
		"OpenCensus Agent exporter host:address")
}

func (opt *serverOption) InitFromViper(v *viper.Viper) {
	opt.storageURL = v.GetString(storageURL)
}

func serveCommand() *cobra.Command {
	vsrv := viper.New()
	cfg := handlers.DefaultAppConfig()
	srvopt := &serverOption{
		storageURL:  "cassandra://kuade_test",
		ocAgentAddr: "127.0.0.1:55678",
		addr:        ":8098",
	}
	serveCommand := &cobra.Command{
		Use:   "serve",
		Short: "start our application server",
		Long:  "serve is used to start kuade application server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// initialize
			storage.DefaultURLMux().InitFromViper(vsrv)
			cfg.InitFromViper(vsrv)
			srvopt.InitFromViper(vsrv)

			ctx := scontext.WithVersion(context.Background(), version.Version)
			// now, open the storage URL
			store, err := storage.OpenStorage(ctx, srvopt.storageURL)
			if err != nil {
				return err
			}

			hd, err := handlers.NewApp(ctx, cfg, assets.Asset, store)
			if err != nil {
				return err
			}

			oc, err := ocagent.NewExporter(
				ocagent.WithInsecure(),
				ocagent.WithAddress(srvopt.ocAgentAddr),
				ocagent.WithServiceName("kuade"))
			if err != nil {
				return err
			}
			defer oc.Stop()

			// setup request logger
			reqlog := requestlog.NewStackdriverLogger(os.Stdout, func(e error) {
				fmt.Println(e)
			})

			webserver := server.New(hd, &server.Options{
				RequestLogger: reqlog,
				TraceExporter: oc,
				Driver:        app.NewDefaultServer()})

			if err = webserver.ListenAndServe(srvopt.addr); err != nil {
				return err
			}

			return nil
		},
	}

	config.AddFlags(vsrv, serveCommand, srvopt.AddFlags, cfg.AddFlags, storage.DefaultURLMux().AddFlags)
	return serveCommand
}
