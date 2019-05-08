package handlers

import (
	"encoding/base64"
	"flag"
	"strings"

	"github.com/spf13/viper"
)

// Config is configuration for Application
type Config struct {
	httpSecure  bool
	sessionKeys [][]byte

	rawsessionKeys string // base64
}

// AddFlags is part of interface config.Configurable
func (c *Config) AddFlags(flagSet *flag.FlagSet) {
	flagSet.Bool(
		"http-secure",
		c.httpSecure,
		"is the HTTP connection served secure? Direct or undirectly using proxy")
	flagSet.String(
		"session-keys",
		c.rawsessionKeys,
		"Session key to be used to encrypt and decrypt session cookie")
}

// InitFromViper is part of interface config.Configurable
func (c *Config) InitFromViper(v *viper.Viper) {
	c.rawsessionKeys = v.GetString("session-keys")
	c.httpSecure = v.GetBool("http-secure")
	if c.rawsessionKeys != "" {
		rawKeys := strings.Split(c.rawsessionKeys, ",")
		c.sessionKeys = make([][]byte, len(rawKeys))
		for _, rawkey := range rawKeys {
			if key, err := c.configureSecretKey(strings.Trim(rawkey, " \t")); err == nil {
				c.sessionKeys = append(c.sessionKeys, key)
			}
		}
	}
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
