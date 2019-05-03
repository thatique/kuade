package cassandra

import (
	"github.com/gocql/gocql"
)

// Configuration describes the configuration properties needed to connect to a Cassandra cluster
type Configuration struct {
	Servers              []string
	Keyspace             string
	LocalDC              string
	ConnectionsPerHost   int
	Timeout              time.Duration
	ReconnectInterval    time.Duration
	SocketKeepAlive      time.Duration
	MaxRetryAttempts     int
	ProtoVersion         int
	Consistency          string
	Port                 int
	Authenticator        Authenticator
	DisableAutoDiscovery bool
	TLS                  TLS
}

// Authenticator holds the authentication properties needed to connect to a Cassandra cluster
type Authenticator struct {
	Basic BasicAuthenticator
	// TODO: add more auth types
}

// BasicAuthenticator holds the username and password for a password authenticator for a Cassandra cluster
type BasicAuthenticator struct {
	Username string
	Password string
}

// TLS Config
type TLS struct {
	Enabled                bool
	ServerName             string
	CertPath               string
	KeyPath                string
	CaPath                 string
	EnableHostVerification bool
}

// ApplyDefaults copies settings from source unless its own value is non-zero.
func (c *Configuration) ApplyDefaults(source *Configuration) {
	if c.ConnectionsPerHost == 0 {
		c.ConnectionsPerHost = source.ConnectionsPerHost
	}
	if c.MaxRetryAttempts == 0 {
		c.MaxRetryAttempts = source.MaxRetryAttempts
	}
	if c.Timeout == 0 {
		c.Timeout = source.Timeout
	}
	if c.ReconnectInterval == 0 {
		c.ReconnectInterval = source.ReconnectInterval
	}
	if c.Port == 0 {
		c.Port = source.Port
	}
	if c.Keyspace == "" {
		c.Keyspace = source.Keyspace
	}
	if c.ProtoVersion == 0 {
		c.ProtoVersion = source.ProtoVersion
	}
	if c.SocketKeepAlive == 0 {
		c.SocketKeepAlive = source.SocketKeepAlive
	}
}

// SessionBuilder creates new cassandra.Session
type SessionBuilder interface {
	NewSession() (*gocql.Session, error)
}

// NewSession creates a new Cassandra session
func (c *Configuration) NewSession() (*gocql.Session, error) {
	cluster := c.NewCluster()
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return session, nil
}

// NewCluster creates a new gocql cluster from the configuration
func (c *Configuration) NewCluster() *gocql.ClusterConfig {
	cluster := gocql.NewCluster(c.Servers...)
	cluster.Keyspace = c.Keyspace
	cluster.NumConns = c.ConnectionsPerHost
	cluster.Timeout = c.Timeout
	cluster.ReconnectInterval = c.ReconnectInterval
	cluster.SocketKeepalive = c.SocketKeepAlive
	if c.ProtoVersion > 0 {
		cluster.ProtoVersion = c.ProtoVersion
	}
	if c.MaxRetryAttempts > 1 {
		cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: c.MaxRetryAttempts - 1}
	}
	if c.Port != 0 {
		cluster.Port = c.Port
	}
	cluster.Compressor = gocql.SnappyCompressor{}
	if c.Consistency == "" {
		cluster.Consistency = gocql.LocalOne
	} else {
		cluster.Consistency = gocql.ParseConsistency(c.Consistency)
	}

	fallbackHostSelectionPolicy := gocql.RoundRobinHostPolicy()
	if c.LocalDC != "" {
		fallbackHostSelectionPolicy = gocql.DCAwareRoundRobinPolicy(c.LocalDC)
	}
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(fallbackHostSelectionPolicy, gocql.ShuffleReplicas())

	if c.Authenticator.Basic.Username != "" && c.Authenticator.Basic.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: c.Authenticator.Basic.Username,
			Password: c.Authenticator.Basic.Password,
		}
	}
	if c.TLS.Enabled {
		cluster.SslOpts = &gocql.SslOptions{
			Config: &tls.Config{
				ServerName: c.TLS.ServerName,
			},
			CertPath:               c.TLS.CertPath,
			KeyPath:                c.TLS.KeyPath,
			CaPath:                 c.TLS.CaPath,
			EnableHostVerification: c.TLS.EnableHostVerification,
		}
	}
	// If tunneling connection to C*, disable cluster autodiscovery features.
	if c.DisableAutoDiscovery {
		cluster.DisableInitialHostLookup = true
		cluster.IgnorePeerAddr = true
	}
	return cluster
}

func (c *Configuration) String() string {
	return fmt.Sprintf("%+v", *c)
}
