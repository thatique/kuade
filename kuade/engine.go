package kuade

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shopify/logrus-bugsnag"
	logstash "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/bugsnag/bugsnag-go"
	gorhandlers "github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"

	"github.com/thatique/kuade/configuration"
	"github.com/thatique/kuade/kuade/listener"
	"github.com/thatique/kuade/pkg/uuid"
	webcontext "github.com/thatique/kuade/pkg/web/context"
	"github.com/thatique/kuade/version"
)

type Application interface {
	GetHTTPHandler(context.Context, *configuration.Configuration) (http.Handler, error)

	Shutdown(context.Context) error
}

type Engine struct {
	app    Application
	ctx    context.Context
	config *configuration.Configuration
	server *http.Server
	quitCh chan os.Signal
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

	handler, err := app.GetHTTPHandler(ctx, config)
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
		app:    app,
		ctx:    ctx,
		config: config,
		server: server,
		quitCh: make(chan os.Signal, 1),
	}, nil
}

func (engine *Engine) ListenAndServe() error {
	config := engine.config
	ln, err := listener.NewListener(config.HTTP.Net, config.HTTP.Addr)
	if err != nil {
		return err
	}

	if config.HTTP.TLS.Certificate != "" {
		var tlsMinVersion uint16
		if config.HTTP.TLS.MinimumTLS == "" {
			tlsMinVersion = tls.VersionTLS10
		} else {
			switch config.HTTP.TLS.MinimumTLS {
			case "tls1.0":
				tlsMinVersion = tls.VersionTLS10
			case "tls1.1":
				tlsMinVersion = tls.VersionTLS11
			case "tls1.2":
				tlsMinVersion = tls.VersionTLS12
			default:
				return fmt.Errorf("unknown minimum TLS level '%s' specified for http.tls.minimumtls", config.HTTP.TLS.MinimumTLS)
			}
			webcontext.GetLogger(engine.ctx).Infof("restricting TLS to %s or higher", config.HTTP.TLS.MinimumTLS)
		}

		tlsConf := &tls.Config{
			ClientAuth:               tls.NoClientCert,
			NextProtos:               []string{"h2", "http/1.1"},
			MinVersion:               tlsMinVersion,
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			},
		}

		tlsConf.Certificates = make([]tls.Certificate, 1)
		tlsConf.Certificates[0], err = tls.LoadX509KeyPair(config.HTTP.TLS.Certificate, config.HTTP.TLS.Key)
		if err != nil {
			return err
		}

		if len(config.HTTP.TLS.ClientCAs) != 0 {
			pool := x509.NewCertPool()

			for _, ca := range config.HTTP.TLS.ClientCAs {
				caPem, err := ioutil.ReadFile(ca)
				if err != nil {
					return err
				}

				if ok := pool.AppendCertsFromPEM(caPem); !ok {
					return fmt.Errorf("could not add CA to pool")
				}
			}

			for _, subj := range pool.Subjects() {
				webcontext.GetLogger(engine.ctx).Debugf("CA Subject: %s", string(subj))
			}

			tlsConf.ClientAuth = tls.RequireAndVerifyClientCert
			tlsConf.ClientCAs = pool
		}

		ln = tls.NewListener(ln, tlsConf)
		webcontext.GetLogger(engine.ctx).Infof("listening on %v, tls", ln.Addr())

		if !config.HTTP.Secure {
			config.HTTP.Secure = true
		}
	} else {
		webcontext.GetLogger(engine.ctx).Infof("listening on %v", ln.Addr())
	}

	// setup channel to get notified on SIGTERM and SIGINT signal
	signal.Notify(engine.quitCh, syscall.SIGTERM, syscall.SIGINT)
	serveErr := make(chan error)

	// Start serving in goroutine and listen for stop signal in main thread
	go func() {
		serveErr <- engine.server.Serve(ln)
	}()

	select {
	case err := <-serveErr:
		return err
	case <-engine.quitCh:
		drainTimeout := time.Second * 60
		if config.HTTP.DrainTimeout != 0 {
			drainTimeout = config.HTTP.DrainTimeout
		}
		webcontext.GetLogger(engine.ctx).Info("stopping server gracefully. Draining connections for ", drainTimeout)
		c, cancel := context.WithTimeout(context.Background(), drainTimeout)
		defer cancel()
		// call application shutdown
		engine.app.Shutdown(c)
		return engine.server.Shutdown(c)
	}
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

		ctx = webcontext.WithValues(ctx, config.Log.Fields)
		ctx = webcontext.WithLogger(ctx, webcontext.GetLogger(ctx, fields...))
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
