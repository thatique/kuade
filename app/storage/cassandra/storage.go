package cassandra

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/spf13/viper"
	"github.com/syaiful6/sersan"
	"github.com/thatique/kuade/app/storage"
	"github.com/thatique/kuade/app/storage/driver"
	"github.com/thatique/kuade/pkg/queue"
)

func init() {
	storage.DefaultURLMux().RegisterStorage(Scheme, &URLOpener{option: NewDefaultOptions()})
}

const Scheme = "cassandra"

// URLOpener opens Cassandra URLs like "cassandra://keyspace".
type URLOpener struct {
	option *Option
}

// OpenStorageURL opens the Cassandra keyspace with the same name as the URL's host.
func (opener *URLOpener) OpenStorageURL(ctx context.Context, u *url.URL) (*storage.Store, error) {
	for param := range u.Query() {
		return nil, fmt.Errorf("open storage %v: invalid query parameter %q", u, param)
	}
	return OpenStorage(ctx, u.Host, opener.option)
}

// AddFlags adds CLI flags for configuring cassandra storage
func (opener *URLOpener) AddFlags(flagSet *flag.FlagSet) {
	opener.option.AddFlags(flagSet)
}

// InitFromViper initializes configuring cassandra storage with properties from spf13/viper.
func (opener *URLOpener) InitFromViper(v *viper.Viper) {
	opener.option.InitFromViper(v)
}

// OpenStorage returns a *storage.Store backed by Cassandra. See the
// package documentation for an example.
func OpenStorage(ctx context.Context, keyspace string, opt *Option) (*storage.Store, error) {
	s, err := openStorage(ctx, keyspace, opt)
	if err != nil {
		return nil, err
	}

	return storage.New(s), nil
}

func openStorage(ctx context.Context, keyspace string, opt *Option) (*Storage, error) {
	config := opt.GetConfig(keyspace)
	session, err := config.NewSession()
	if err != nil {
		return nil, err
	}

	s := &Storage{
		session: session,
		queue:   queue.NewBoundedQueue(opt.capacity, nil),
	}
	s.queue.StartConsumer(opt.numWorker, func(item interface{}) {
		value := item.(asyncQuery)
		value.Fire(session)
	})

	return s, nil
}

type asyncQuery interface {
	Fire(*gocql.Session) error
}

type asyncQueryFunc func(*gocql.Session) error

func (f asyncQueryFunc) Fire(sess *gocql.Session) error {
	return f(sess)
}

// Storage is cassandra implementation of storage driver
type Storage struct {
	session *gocql.Session
	// used to execute query asyncronously
	queue *queue.BoundedQueue
	// set up when first request
	users *userStore
}

// GetUserStore return driver.UserStore implementation backed by cassandra
func (s *Storage) GetUserStore() (driver.UserStore, error) {
	if s.users == nil {
		s.users = &userStore{session: s.session}
	}

	return s.users, nil
}

// GetSessionStore return driver.GetSessionStore implementation backed by cassandra
func (s *Storage) GetSessionStore() (sersan.Storage, error) {
	return &sessionStore{session: s.session, queue: s.queue}, nil
}

func (s *Storage) Close() error {
	s.queue.Stop()
	s.session.Close()

	return nil
}

// Option for connection to Cassandra server
type Option struct {
	Configuration

	servers string

	capacity  int
	numWorker int
}

// NewDefaultOption return default Option
func NewDefaultOptions() *Option {
	return &Option{
		Configuration: Configuration{
			TLS: TLS{
				Enabled:                false,
				EnableHostVerification: true,
			},
			MaxRetryAttempts:   3,
			Keyspace:           "kuade_v1_test",
			ProtoVersion:       4,
			ConnectionsPerHost: 2,
			ReconnectInterval:  60 * time.Second,
		},
		servers:   "127.0.0.1",
		numWorker: 50,
		capacity:  2000,
	}
}

// GetConfig get Configuration backed by this configuration populated using viper
// package
func (opt *Option) GetConfig(keyspace string) *Configuration {
	opt.Servers = strings.Split(opt.servers, ",")
	opt.Keyspace = keyspace
	return &opt.Configuration
}

// AddFlags add cli flag for this option
func (opt *Option) AddFlags(flagSet *flag.FlagSet) {
	flagSet.Int(
		"cassandra-connections-per-host",
		opt.ConnectionsPerHost,
		"The number of Cassandra connections from a single backend instance")
	flagSet.Int(
		"cassandra-max-retry-attempts",
		opt.MaxRetryAttempts,
		"The number of attempts when reading from Cassandra")
	flagSet.Duration(
		"cassandra-timeout",
		opt.Timeout,
		"Timeout used for queries. A Timeout of zero means no timeout")
	flagSet.Duration(
		"cassandra-reconnect-interval",
		opt.ReconnectInterval,
		"Reconnect interval to retry connecting to downed hosts")
	flagSet.String(
		"cassandra-servers",
		opt.servers,
		"The comma-separated list of Cassandra servers")
	flagSet.Int(
		"cassandra-port",
		opt.Port,
		"The port for cassandra")
	flagSet.String(
		"cassandra-local-dc",
		opt.LocalDC,
		"The name of the Cassandra local data center for DC Aware host selection")
	flagSet.String(
		"cassandra-consistency",
		opt.Consistency,
		"The Cassandra consistency level, e.g. ANY, ONE, TWO, THREE, QUORUM, ALL, LOCAL_QUORUM, EACH_QUORUM, LOCAL_ONE (default LOCAL_ONE)")
	flagSet.Int(
		"cassandra-proto-version",
		opt.ProtoVersion,
		"The Cassandra protocol version")
	flagSet.Duration(
		"cassandra-socket-keep-alive",
		opt.SocketKeepAlive,
		"Cassandra's keepalive period to use, enabled if > 0")
	flagSet.String(
		"cassandra-username",
		opt.Authenticator.Basic.Username,
		"Username for password authentication for Cassandra")
	flagSet.String(
		"cassandra-password",
		opt.Authenticator.Basic.Password,
		"Password for password authentication for Cassandra")
	flagSet.Bool(
		"cassandra-tls",
		opt.TLS.Enabled,
		"Enable TLS")
	flagSet.String(
		"cassandra-tls-cert",
		opt.TLS.CertPath,
		"Path to TLS certificate file")
	flagSet.String(
		"cassandra-tls-key",
		opt.TLS.KeyPath,
		"Path to TLS key file")
	flagSet.String(
		"cassandra-tls-ca",
		opt.TLS.CaPath,
		"Path to TLS CA file")
	flagSet.String(
		"cassandra-tls-server-name",
		opt.TLS.ServerName,
		"Override the TLS server name")
	flagSet.Bool(
		"cassandra-tls-verify-host",
		opt.TLS.EnableHostVerification,
		"Enable (or disable) host key verification")
	flagSet.Int(
		"cassandra-queue-capacity",
		opt.capacity,
		"size of the queue to process async query")
	flagSet.Int(
		"cassandra-queue-worker",
		opt.numWorker,
		"Number of worker used to execute async query")
}

// initFromViper initialize option using viper variable
func (opt *Option) InitFromViper(v *viper.Viper) {
	opt.ConnectionsPerHost = v.GetInt("cassandra-connections-per-host")
	opt.MaxRetryAttempts = v.GetInt("cassandra-max-retry-attempts")
	opt.Timeout = v.GetDuration("cassandra-timeout")
	opt.ReconnectInterval = v.GetDuration("cassandra-reconnect-interval")
	opt.servers = strings.Replace(v.GetString("cassandra-servers"), " ", "", -1)
	opt.Port = v.GetInt("cassandra-port")
	opt.LocalDC = v.GetString("cassandra-local-dc")
	opt.Consistency = v.GetString("cassandra-consistency")
	opt.ProtoVersion = v.GetInt("cassandra-proto-version")
	opt.SocketKeepAlive = v.GetDuration("cassandra-socket-keep-alive")
	opt.Authenticator.Basic.Username = v.GetString("cassandra-username")
	opt.Authenticator.Basic.Password = v.GetString("cassandra-password")
	opt.TLS.Enabled = v.GetBool("cassandra-tls")
	opt.TLS.CertPath = v.GetString("cassandra-tls-cert")
	opt.TLS.KeyPath = v.GetString("cassandra-tls-key")
	opt.TLS.CaPath = v.GetString("cassandra-tls-ca")
	opt.TLS.ServerName = v.GetString("cassandra-tls-server-name")
	opt.TLS.EnableHostVerification = v.GetBool("cassandra-tls-verify-host")
	opt.numWorker = v.GetInt("cassandra-queue-worker")
}
