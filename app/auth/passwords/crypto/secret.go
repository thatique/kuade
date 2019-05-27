package secret

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"time"

	"github.com/thatique/kuade/app/model"
)

type SecretToken struct {
	key [32]byte
}

func (s *SecretToken) Create(ctx context.Context, cred *model.Credentials) (string, error) {
	return s.createTokenWithTimestamp(ctx, cred, numDays(time.Now().UTC()))
}

func (s *SecretToken) createTokenWithTimestamp(ctx context.Context, cred *model.Credentials, timestamp uint64) (string, error) {
}

func (s *SecretToken) getMac(cred *model.Credentials, timestamp uint64) string {
	hsh := hmac.New(sha256.New, h.key[:])
	hsh.Write(makeHashValue(cred, timestamp)
	return base64.RawURLEncoding.EncodeToString(hsh.Sum(nil))
}

// timestampt+userID+password+lastSignin
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

func numDays(dt time.Time) uint64 {
	var start = time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC)
	diff := dt.Sub(start)

	return uint64(diff.Hours()/24)
}
