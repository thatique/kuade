package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

type IfRequestMiddleware struct {
	predicate   func(*http.Request) bool
	middlewares []mux.MiddlewareFunc
}

func NewIfRequestMiddleware(middlewares []mux.MiddlewareFunc, predicate func(*http.Request) bool) *IfRequestMiddleware {
	return &IfRequestMiddleware{predicate: predicate, middlewares: middlewares,}
}

func (mw *IfRequestMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if mw.predicate(req) {
			for i := len(mw.middlewares) - 1; i >= 0; i-- {
				next = mw.middlewares[i](next)
			}
		}
		next.ServeHTTP(w, req)
	})
}
