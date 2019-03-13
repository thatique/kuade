package kuade

import (
	"context"
	"net/http"

	"github.com/bugsnag/bugsnag-go"
	logstash "github.com/bshuster-repo/logrus-logstash-hook"
	gorhandlers "github.com/gorilla/handlers"
	"github.com/Shopify/logrus-bugsnag"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/thatique/kuade/configuration"
	"github.com/thatique/kuade/version"
	"github.com/thatique/kuade/pkg/queue"
	"github.com/thatique/kuade/pkg/uuid"
	webcontext "github.com/thatique/kuade/pkg/web/context"
)

type Application interface {
	GetHTTPHandler(context.Context, *configuration.Configuration) (http.Handler, error)
}

type AppFeatures struct {
	JobChan chan<- Job
}

type Engine struct {
	config *configuration.Configuration
	server *http.Server
}

func NewEngine(app Application, config *configuration.Configuration) (*Engine, error) {
	ctx := webcontext.WithVersion(webcontext.Background(), version.Version)

	var err error
	ctx, err = configureLogging(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error configuring logger: %v", err)
	}

	configureBugsnag(config)

	// inject a logger into the uuid library. warns us if there is a problem
	// with uuid generation under low entropy.
	uuid.Loggerf = webcontext.GetLogger(ctx).Warnf

	appHandler, err := app.GetHTTPHandler(ctx, config)
	if err != nil {
		return nil, err
	}

	if !config.Log.AccessLog.Disabled {
		handler = gorhandlers.CombinedLoggingHandler(os.Stdout, handler)
	}

	server := &http.Server{
		Handler: handler,
	}

	return &Engine{
		config: config,
		server: server,
	}, nil
}

// panicHandler add an HTTP handler to web app. The handler recover the happening
// panic. logrus.Panic transmits panic message to pre-config log hooks, which is
// defined in config.yml.
func panicHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Panic(fmt.Sprintf("%v", err))
			}
		}()
		handler.ServeHTTP(w, r)
	})
}

// configureBugsnag configures bugsnag reporting, if enabled
func configureBugsnag(config *configuration.Configuration) {
	if config.Reporting.Bugsnag.APIKey == "" {
		return
	}

	bugsnagConfig := bugsnag.Configuration{
		APIKey: config.Reporting.Bugsnag.APIKey,
	}
	if config.Reporting.Bugsnag.ReleaseStage != "" {
		bugsnagConfig.ReleaseStage = config.Reporting.Bugsnag.ReleaseStage
	}
	if config.Reporting.Bugsnag.Endpoint != "" {
		bugsnagConfig.Endpoint = config.Reporting.Bugsnag.Endpoint
	}
	bugsnag.Configure(bugsnagConfig)

	// configure logrus bugsnag hook
	hook, err := logrus_bugsnag.NewBugsnagHook()
	if err != nil {
		log.Fatalln(err)
	}

	log.AddHook(hook)
}

// configureLogging prepares the context with a logger using the
// configuration.
func configureLogging(ctx context.Context, config *configuration.Configuration) (context.Context, error) {
	log.SetLevel(logLevel(config.Log.Level))

	formatter := config.Log.Formatter
	if formatter == "" {
		formatter = "text" // default formatter
	}

	switch formatter {
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	case "text":
		log.SetFormatter(&log.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	case "logstash":
		log.SetFormatter(&logstash.LogstashFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	default:
		// just let the library use default on empty string.
		if config.Log.Formatter != "" {
			return ctx, fmt.Errorf("unsupported logging formatter: %q", config.Log.Formatter)
		}
	}

	if config.Log.Formatter != "" {
		log.Debugf("using %q logging formatter", config.Log.Formatter)
	}

	if len(config.Log.Fields) > 0 {
		// build up the static fields, if present.
		var fields []interface{}
		for k := range config.Log.Fields {
			fields = append(fields, k)
		}

		ctx = scontext.WithValues(ctx, config.Log.Fields)
		ctx = scontext.WithLogger(ctx, scontext.GetLogger(ctx, fields...))
	}

	return ctx, nil
}

func logLevel(level configuration.Loglevel) log.Level {
	l, err := log.ParseLevel(string(level))
	if err != nil {
		l = log.InfoLevel
		log.Warnf("error parsing level %q: %v, using %q	", level, err, l)
	}

	return l
}
