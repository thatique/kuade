package handlers

import (
	"net/http"
)

type RequestSizeLimit struct {
	maxBodySize int64
}

func NewRequestLimiter(maxBodySize int64) *RequestSizeLimit {
	return &RequestSizeLimit{maxBodySize: maxBodySize}
}

func (limiter *RequestSizeLimit) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Restricting read data to a given maximum length
		r.Body = http.MaxBytesReader(w, r.Body, limiter.maxBodySize)
		next.ServeHTTP(w, r)
	})
}
