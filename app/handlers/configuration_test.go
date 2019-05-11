package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"
	"testing"
)

func generateBase64Key() string {
	k := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		panic(err)
	}
	return "base64:" + base64.StdEncoding.EncodeToString(k)
}

func TestCommaSepatedKeysParsed(t *testing.T) {
	var c *Config
	var b strings.Builder
	for i := 0; i < 4; i++ {
		b.WriteString(generateBase64Key())
		b.WriteRune(',')
	}
	rawKeys := strings.Trim(b.String(), ",")
	keys := c.configureSecretKeys(rawKeys)
	if len(keys) != 4 {
		t.Fatalf("expected configure secrets returns %d keys length. it's has %d instead. %v", 4, len(keys), keys)
	}
}
