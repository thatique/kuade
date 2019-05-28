package cryptotoken

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/thatique/kuade/app/model"
)

type CryptoToken struct {
	key               [32]byte
	resetTimeoutHours uint64
}

func New(key [32]byte, expiry uint64) *CryptoToken {
	return &CryptoToken{key: key, resetTimeoutHours: expiry}
}

func (s *CryptoToken) Create(ctx context.Context, cred *model.Credentials) (string, error) {
	return s.createTokenWithTimestamp(ctx, cred, numHours(time.Now().UTC()))
}

func (s *CryptoToken) Check(ctx context.Context, cred *model.Credentials, token string) bool {
	xs := strings.SplitN(token, "-", 2)
	if len(xs) != 2 {
		return false
	}

	ts, err := strconv.ParseUint(xs[0], 16, 64)
	if err != nil {
		return false
	}

	signature, err := base64.RawURLEncoding.DecodeString(xs[1])
	if err != nil {
		return false
	}
	if subtle.ConstantTimeCompare(signature, s.getMac(cred, ts)) != 1 {
		return false
	}

	// Check the timestamp is within limit.
	diff := numHours(time.Now().UTC()) - ts
	if diff > s.resetTimeoutHours {
		return false
	}
	return true
}

func (s *CryptoToken) createTokenWithTimestamp(ctx context.Context, cred *model.Credentials, timestamp uint64) (string, error) {
	return fmt.Sprintf("%x-%s", timestamp, base64.RawURLEncoding.EncodeToString(s.getMac(cred, timestamp))), nil
}

func (s *CryptoToken) getMac(cred *model.Credentials, timestamp uint64) []byte {
	hsh := hmac.New(sha256.New, s.key[:])
	hsh.Write(makeHashValue(cred, timestamp))
	return hsh.Sum(nil)
}

// timestamp+userID+password+lastSignin
func makeHashValue(cred *model.Credentials, timestamp uint64) []byte {
	var b bytes.Buffer
	var bt [8]byte
	binary.BigEndian.PutUint64(bt[:], timestamp)
	b.Write(bt[:])
	var it [8]byte
	cred.UserID.MarshalTo(it[:])
	b.Write(it[:])
	b.Write(cred.Password)
	var lt [8]byte
	binary.BigEndian.PutUint64(lt[:], uint64(cred.LastSignin.Unix()))
	b.Write(lt[:])

	return b.Bytes()
}

func numHours(dt time.Time) uint64 {
	var start = time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC)
	diff := dt.Sub(start)

	return uint64(diff.Hours())
}
