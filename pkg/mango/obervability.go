package mango

import (
	"context"
	"sync"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
)

var (
	keyMethod, _ = tag.NewKey("method")
	keyStatus, _ = tag.NewKey("status")
	keyError, _  = tag.NewKey("error")
)

var mLatencyMs = stats.Float64("latency", "the Latency in miliseconds", "ms")

var latencyDistribution = view.Distribution(
	0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 1.5, 2, 2.5, 5, 10, 25, 50, 100, 200,
	400, 600, 800, 1000, 1500, 2000, 2500, 5000, 10000, 20000, 40000, 100000, 200000, 500000, 1000000)

var allViews = []*view.View{
	{
		Name:        "mongo/client/latency",
		Description: "The latency of the various calls",
		Aggregation: latencyDistribution,
		TagKeys:     []tag.Key{keyMethod, keyStatus, keyError},
	},
	{
		Name:        "mongo/client/calls",
		Description: "The various calls",
		Measure:     mLatencyMs,
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{keyMethod, keyStatus, keyError},
	},
}

func RegisterAllViews() error {
	return view.Register(allViews...)
}

func UnregisterAllViews() {
	view.Unregister(allViews...)
}

type spanWithMetrics struct {
	startTime time.Time
	method    string
	lastErr   error
	span      *trace.Span
	endOnce   sync.Once
}

func roundtripTrackingSpan(ctx context.Context, methodName string, traceOpts ...trace.StartOption) (context.Context, *spanWithMetrics) {
	ctx, span := trace.StartSpan(ctx, methodName, traceOpts...)
	return ctx, &spanWithMetrics{span: span, startTime: time.Now(), method: methodName}
}

func (swm *spanWithMetrics) setError(err error) {
	if err != nil {
		swm.span.SetStatus(trace.Status{Code: trace.StatusCodeUnknown, Message: err.Error()})
	}

	swm.lastErr = err
}

func (swm *spanWithMetrics) end(ctx context.Context) {
	swm.endOnce.Do(func() {
		if err := swm.lastErr; err == nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyMethod, swm.method), tag.Upsert(keyStatus, "OK"))
		} else {
			ctx, _ = tag.New(ctx, tag.Upsert(keyMethod, swm.method), tag.Upsert(keyError, err.Error()))
		}

		latencyMs := float64(time.Now().Sub(swm.startTime)) / 1e6
		stats.Record(ctx, mLatencyMs.M(latencyMs))
		swm.span.End()
	})
}
