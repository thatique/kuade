package handlers

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/syaiful6/sersan"
	redistore "github.com/syaiful6/sersan/redis"

	"github.com/thatique/kuade/app"
	"github.com/thatique/kuade/app/auth"
	"github.com/thatique/kuade/app/service"
	"github.com/thatique/kuade/configuration"
	webcontext "github.com/thatique/kuade/pkg/web/context"
	"github.com/thatique/kuade/pkg/web/handlers"
)

type App struct {
	context.Context
	// assets function
	asset func(string) ([]byte, error)

	authSess *auth.Session
	router   *mux.Router
	service  *service.Service
}

func NewApp(asset func(string) ([]byte, error)) app.Application {
	return &App{asset: asset}
}

func (app *App) GetService() *service.Service {
	return app.service
}

func (app *App) GetHTTPHandler(ctx context.Context, conf *configuration.Configuration) (http.Handler, error) {
	svc, err := service.NewService(conf)
	if err != nil {
		return nil, err
	}

	app.Context = ctx

	router := routerWithPrefix(conf.HTTP.Prefix)
	app.service = svc
	app.router = router

	userStorage, err := svc.Storage.GetUserStorage()
	if err != nil {
		return nil, err
	}

	app.authSess = auth.NewUserSession(userStorage)

	sessionState, err := app.configureSersan(conf)
	if err != nil {
		return nil, err
	}

	webMiddlewares := handlers.NewIfRequestMiddleware([]mux.MiddlewareFunc{
		sersan.SessionMiddleware(sessionState),
		app.authSess.Middleware,
		csrf.Protect([]byte(conf.HTTP.Secret), csrf.Secure(conf.HTTP.Secure)),
	}, isNotApiRoute)
	// middleware
	app.router.Use(webMiddlewares.Middleware)

	// static files
	app.router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(handlers.NewStaticFS("assets/static", app.asset))))

	return app, nil
}

func (app *App) Shutdown(ctx context.Context) error {
	app.service.Quit()
	return nil
}

// configureSecret creates a random secret if a secret wasn't included in the
// configuration.
func (app *App) configureSecret(configuration *configuration.Configuration) {
	if configuration.HTTP.Secret == "" {
		var secretBytes [32]byte
		if _, err := cryptorand.Read(secretBytes[:]); err != nil {
			panic(fmt.Sprintf("could not generate random bytes for HTTP secret: %v", err))
		}
		configuration.HTTP.Secret = string(secretBytes[:])
		webcontext.GetLogger(app).Warn("No HTTP secret provided - generated random secret.")
	}
}

func (app *App) configureSersan(conf *configuration.Configuration) (*sersan.ServerSessionState, error) {
	sersanstore, err := redistore.NewRediStore(app.service.Redis)
	if err != nil {
		return nil, err
	}

	sessionKeys := createSecretKeys(conf.HTTP.SessionKeys...)
	if len(sessionKeys) == 0 {
		return nil, errors.New("http session keys must not be empty.")
	}
	sessionstate := sersan.NewServerSessionState(sersanstore, sessionKeys...)
	sessionstate.AuthKey = auth.UserSessionKey
	sessionstate.Options.Secure = conf.HTTP.Secure

	return sessionstate, nil
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // ensure that request body is always closed.

	// Prepare the context with our own little decorations.
	ctx := app.context(w, r)
	ctx = webcontext.WithRequest(ctx, r)
	ctx, w = webcontext.WithResponseWriter(ctx, w)
	ctx = webcontext.WithLogger(ctx, webcontext.GetRequestLogger(ctx))
	// sync the context
	r = r.WithContext(ctx)

	for headerName, headerValues := range app.service.Config.HTTP.Headers {
		for _, value := range headerValues {
			w.Header().Add(headerName, value)
		}
	}

	defer func() {
		status, ok := ctx.Value("http.response.status").(int)
		if ok && status >= 200 && status <= 399 {
			webcontext.GetResponseLogger(r.Context()).Infof("response completed")
		}
	}()

	app.router.ServeHTTP(w, r)
}

// context constructs the context object for the application. This only be
// called once per request.
func (app *App) context(w http.ResponseWriter, r *http.Request) context.Context {
	ctx := r.Context()
	ctx = webcontext.WithVars(ctx, r)
	ctx = webcontext.WithLogger(ctx, webcontext.GetLogger(ctx,
		"vars.name",
		"vars.uuid"))

	return &Context{
		App:     app,
		Context: ctx,
	}
}

func routerWithPrefix(prefix string) *mux.Router {
	rootRouter := mux.NewRouter()
	router := rootRouter
	if prefix != "" {
		router = router.PathPrefix(prefix).Subrouter()
	}

	router.StrictSlash(true)

	return rootRouter
}

func createSecretKeys(keyPairs ...string) [][]byte {
	xs := make([][]byte, len(keyPairs))
	var (
		err error
		key []byte
	)
	for _, s := range keyPairs {
		if strings.HasPrefix(s, "base64:") {
			key, err = base64.StdEncoding.DecodeString(strings.TrimPrefix(s, "base64:"))
			if err != nil {
				continue
			}
			xs = append(xs, key)
		} else {
			xs = append(xs, []byte(s))
		}
	}
	return xs
}

func isNotApiRoute(r *http.Request) bool {
	return !strings.HasPrefix(r.URL.Path, "/api")
}
