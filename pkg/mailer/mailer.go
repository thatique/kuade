package mailer

import (
	"context"
	"net/mail"
	"net/url"
	"strings"

	"github.com/emersion/go-message"
	"github.com/thatique/kuade/pkg/mailer/driver"
	"github.com/thatique/kuade/pkg/kerr"
	"github.com/thatique/kuade/pkg/metrics"
	"github.com/thatique/kuade/pkg/openurl"
)

const pkgName = "github.com/thatique/kuade/pkg/mailer"

var (
	latencyMeasure  = metrics.LatencyMeasure(pkgName)

	// OpenCensusViews are predefined views for OpenCensus metrics.
	// The views include counts and latency distributions for API method calls.
	// See the example at https://godoc.org/go.opencensus.io/stats/view for usage.
	OpenCensusViews = metrics.Views(pkgName, latencyMeasure)
)

type Transport struct {
	transport driver.Transport
	tracer    *metrics.Tracer
}

func NewMailer(transport driver.Transport) *Transport {
	return &Transport{
		transport: transport,
		tracer: &metrics.Tracer{
			Package:        "github.com/thatique/kuade/pkg/mailer",
			Provider:       metrics.ProviderName(transport),
			LatencyMeasure: latencyMeasure,
		},
	}
}

func (m *Transport) Open(ctx context.Context) (err error) {
	ctx = m.tracer.Start(ctx, "Open")
	defer func() { m.tracer.End(ctx, err) }()

	err = m.transport.Open(ctx)
	if err != nil {
		err = wrapError(m, err)
	}
	return
}

func (m *Transport) Close(ctx context.Context) (err error) {
	ctx = m.tracer.Start(ctx, "Close")
	defer func() { m.tracer.End(ctx, err) }()

	err = m.transport.Close(ctx)
	if err != nil {
		err = wrapError(m, err)
	}
	return
}

func (m *Transport) SendMessages(ctx context.Context, messages []*message.Entity) (n int, err error) {
	ctx = m.tracer.Start(ctx, "SendMessages")
	defer func() { m.tracer.End(ctx, err) }()

	n, err = m.transport.SendMessages(ctx, messages)
	if err != nil {
		err = wrapError(m, err)
	}
	return
}

func wrapError(m *Transport, err error) error {
	if kerr.DoNotWrap(err) {
		return err
	}
	return kerr.New(m.transport.ErrorCode(err), err, 2, "secrets")
}

type TransportURLOpener interface {
	OpenTransportURL(ctx context.Context, u *url.URL) (*Transport, error)
}

type URLMux struct {
	schemes openurl.SchemaMap
}

// RegisterKeeper registers the opener with the given scheme. If an opener
// already exists for the scheme, RegisterKeeper panics.
func (mux *URLMux) RegisterTransport(scheme string, opener TransportURLOpener) {
	mux.schemes.Register("transport", "Mailer", scheme, opener)
}

func (mux *URLMux) OpenTransport(ctx context.Context, urlstr string) (*Transport, error) {
	opener, u, err := mux.schemes.FromString("Mailer", urlstr)
	if err != nil {
		return nil, err
	}
	return opener.(TransportURLOpener).OpenTransportURL(ctx, u)
}

// OpenTransportURL dispatches the URL to the opener that is registered with the
// URL's scheme. OpenTransportURL is safe to call from multiple goroutines.
func (mux *URLMux) OpenTransportURL(ctx context.Context, u *url.URL) (*Transport, error) {
	opener, err := mux.schemes.FromURL("Mailer", u)
	if err != nil {
		return nil, err
	}
	return opener.(TransportURLOpener).OpenTransportURL(ctx, u)
}

var defaultURLMux = new(URLMux)

// DefaultURLMux returns the URLMux used by OpenTransport.
//
// Driver packages can use this to register their TransportURLOpener on the mux.
func DefaultURLMux() *URLMux {
	return defaultURLMux
}

// OpenTransport opens the Keeper identified by the URL given.
// See the URLOpener documentation in provider-specific subpackages for
// details on supported URL formats
func OpenTransport(ctx context.Context, urlstr string) (*Transport, error) {
	return defaultURLMux.OpenTransport(ctx, urlstr)
}

// FormatAddress format an array of `mail.Address` as a list of comma-separated
// addresses of the form "Gogh Fir <gf@example.com>" or "foo@example.com".
func FormatAddressList(xs []*mail.Address) string {
	formatted := make([]string, len(xs))
	for i, a := range xs {
		formatted[i] = a.String()
	}

	return strings.Join(formatted, ", ")
}
