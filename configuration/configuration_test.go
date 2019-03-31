package configuration

import (
	"bytes"
	"net/http"
	"os"
	"testing"
	"time"

	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

// Hook up gocheck into the "go test" runner
func Test(t *testing.T) { TestingT(t) }

var configStruct = Configuration{
	Version:  "0.1",
	DataPath: "/data",
	Log: struct {
		AccessLog struct {
			Disabled bool `yaml:"disabled,omitempty"`
		} `yaml:"accesslog,omitempty"`
		Level     Loglevel               `yaml:"level,omitempty"`
		Formatter string                 `yaml:"formatter,omitempty"`
		Fields    map[string]interface{} `yaml:"fields,omitempty"`
		Hooks     []LogHook              `yaml:"hooks,omitempty"`
	}{
		Level:  "info",
		Fields: map[string]interface{}{"environment": "test"},
	},

	Reporting: Reporting{
		Bugsnag: BugsnagReporting{
			APIKey: "BugsnagApiKey",
		},
	},

	HTTP: struct {
		Addr        string   `yaml:"addr,omitempty"`
		Net         string   `yaml:"net,omitempty"`
		Host        string   `yaml:"host,omitempty"`
		Prefix      string   `yaml:"prefix,omitempty"`
		Secret      string   `yaml:"secret,omitempty"`
		SessionKeys []string `yaml:"sessionkeys,omitempty"`
		Secure      bool     `yaml:"secure,omitempty"`
		TLS         struct {
			Certificate string   `yaml:"certificate,omitempty"`
			Key         string   `yaml:"key,omitempty"`
			ClientCAs   []string `yaml:"clientcas,omitempty"`
			MinimumTLS  string   `yaml:"minimumtls,omitempty"`
		} `yaml:"tls,omitempty"`
		DrainTimeout time.Duration `yaml:"draintimeout,omitempty"`
		Headers      http.Header   `yaml:"headers,omitempty"`

		Debug struct {
			// Addr specifies the bind address for the debug server.
			Addr string `yaml:"addr,omitempty"`
			// Prometheus configures the Prometheus telemetry endpoint.
			Prometheus struct {
				Enabled bool   `yaml:"enabled,omitempty"`
				Path    string `yaml:"path,omitempty"`
			} `yaml:"prometheus,omitempty"`
		} `yaml:"debug,omitempty"`
	}{
		Addr: "localhost",
		TLS: struct {
			Certificate string   `yaml:"certificate,omitempty"`
			Key         string   `yaml:"key,omitempty"`
			ClientCAs   []string `yaml:"clientcas,omitempty"`
			MinimumTLS  string   `yaml:"minimumtls,omitempty"`
		}{
			ClientCAs: []string{"/path/to/ca.pem"},
		},
		Headers: http.Header{
			"X-Content-Type-Options": []string{"nosniff"},
		},
	},

	Redis: Redis{
		Addr:     "localhost",
		Password: "secret",
		DB:       1,
	},

	Queue: Queue{
		MaxWorkers: 3,
		MaxQueue:   100,
	},

	Storage: Storage{
		"mongodb": Parameters{
			"url": "mongodb://localhost:2703/db",
		},
	},
}

// configYamlV0_1 is a Version 0.1 yaml document representing configStruct
var configYamlV0_1 = `
version: 0.1
datapath: /data
log:
  level: info
  fields:
    environment: test
reporting:
  bugsnag:
    apiKey: BugsnagApiKey
http:
  addr: localhost
  tls:
    clientcas:
      - /path/to/ca.pem
  headers:
    X-Content-Type-Options: [nosniff]
redis:
  addr: localhost
  password: secret
  db: 1
queue:
  maxworkers: 3
  maxqueue: 100
storage:
  mongodb:
    url: mongodb://localhost:2703/db
`

type ConfigSuite struct {
	expectedConfig *Configuration
}

var _ = Suite(new(ConfigSuite))

func (suite *ConfigSuite) SetUpTest(c *C) {
	os.Clearenv()
	suite.expectedConfig = copyConfig(configStruct)
}

// TestMarshalRoundtrip validates that configStruct can be marshaled and
// unmarshaled without changing any parameters
func (suite *ConfigSuite) TestMarshalRoundtrip(c *C) {
	configBytes, err := yaml.Marshal(suite.expectedConfig)
	c.Assert(err, IsNil)
	config, err := Parse(bytes.NewReader(configBytes))
	c.Log(string(configBytes))
	c.Assert(err, IsNil)
	c.Assert(config, DeepEquals, suite.expectedConfig)
}

// TestParseSimple validates that configYamlV0_1 can be parsed into a struct
// matching configStruct
func (suite *ConfigSuite) TestParseSimple(c *C) {
	config, err := Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, IsNil)
	c.Assert(config, DeepEquals, suite.expectedConfig)
}

func copyConfig(config Configuration) *Configuration {
	configCopy := new(Configuration)

	configCopy.DataPath = config.DataPath
	configCopy.Version = MajorMinorVersion(config.Version.Major(), config.Version.Minor())

	configCopy.Log = config.Log
	configCopy.Log.Fields = make(map[string]interface{}, len(config.Log.Fields))
	for k, v := range config.Log.Fields {
		configCopy.Log.Fields[k] = v
	}

	configCopy.HTTP = config.HTTP
	configCopy.HTTP.Headers = make(http.Header)
	for k, v := range config.HTTP.Headers {
		configCopy.HTTP.Headers[k] = v
	}

	configCopy.Queue = config.Queue
	configCopy.Redis = config.Redis
	configCopy.Storage =  Storage{config.Storage.Type(): Parameters{}}
	for k, v := range config.Storage.Parameters() {
		configCopy.Storage.setParameter(k, v)
	}

	return configCopy
}
