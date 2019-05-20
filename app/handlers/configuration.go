package handlers

import (
	"encoding/base64"
	"flag"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config is configuration for Application
type Config struct {
	httpSecure     bool
	sessionKeys    [][]byte
	configFile     string
	rawsessionKeys string // base64
	logLevel       string
}

// DefaultAppConfig return the default Config
func DefaultAppConfig() *Config {
	return &Config{
		httpSecure: false,
		logLevel:   "info",
	}
}

// AddFlags is part of interface config.Configurable
func (c *Config) AddFlags(flagSet *flag.FlagSet) {
	flagSet.String(
		"config-file",
		c.configFile,
		"Configuration file in JSON, TOML, YAML, HCL, or Java properties formats (default none). See spf13/viper for precedence")
	flagSet.Bool(
		"http-secure",
		c.httpSecure,
		"is the HTTP connection served secure? Direct or undirectly using proxy")
	flagSet.String(
		"session-keys",
		c.rawsessionKeys,
		"Session key to be used to encrypt and decrypt session cookie")
	flagSet.String(
		"log-level",
		c.logLevel,
		"Minimal allowed log level. For more levels see https://github.com/uber-go/zap")
}

func (c *Config) tryLoadConfigFile(v *viper.Viper) error {
	if file := v.GetString("config-file"); file != "" {
		v.SetConfigFile(file)
		err := v.ReadInConfig()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) NewLogger(conf zap.Config, options ...zap.Option) (*zap.Logger, error) {
	var level zapcore.Level
	err := (&level).UnmarshalText([]byte(c.logLevel))
	if err != nil {
		return nil, err
	}
	conf.Level = zap.NewAtomicLevelAt(level)
	return conf.Build(options...)
}

// InitFromViper is part of interface config.Configurable
func (c *Config) InitFromViper(v *viper.Viper) {
	c.rawsessionKeys = v.GetString("session-keys")
	c.httpSecure = v.GetBool("http-secure")
	if c.rawsessionKeys != "" {
		c.sessionKeys = c.configureSecretKeys(c.rawsessionKeys)
	}
	c.tryLoadConfigFile(v)
	c.logLevel = v.GetString("log-level")
}

func (c *Config) configureSecretKeys(s string) [][]byte {
	rawKeys := strings.Split(s, ",")
	var sessionKeys [][]byte
	for _, rawkey := range rawKeys {
		if key, err := c.configureSecretKey(strings.Trim(rawkey, " \t")); err == nil {
			sessionKeys = append(sessionKeys, key)
		}
	}

	return sessionKeys
}

func (c *Config) configureSecretKey(s string) ([]byte, error) {
	if strings.HasPrefix(s, "base64:") {
		key, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(s, "base64:"))
		if err != nil {
			return nil, err
		}
		return key, nil
	}
	return []byte(s), nil
}
