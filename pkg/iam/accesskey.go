package iam

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/thatique/kuade/pkg/text"
)

const (
	// Minimum length for Minio access key.
	accessKeyMinLen = 3

	// Maximum length for Minio access key.
	// There is no max length enforcement for access keys
	accessKeyMaxLen = 20

	// Minimum length for Minio secret key for both server and gateway mode.
	secretKeyMinLen = 8

	// Maximum secret key length, this is used when autogenerating new access key.
	// There is no max length enforcement for secret keys
	secretKeyMaxLen = 40

	AccessKeyChars = text.ASCII_LOWERCASE + text.ASCII_UPPERCASE + text.DIGITS
	SecretKeyChars = AccessKeyChars + "-_~"
)

// Common errors generated for access and secret key validation.
var (
	ErrInvalidAccessKeyLength = fmt.Errorf("access key must be minimum %v or more characters long", accessKeyMinLen)
	ErrInvalidSecretKeyLength = fmt.Errorf("secret key must be minimum %v or more characters long", secretKeyMinLen)
)

type AccessKey struct {
	AccessKey    string    `xml:"AccessKeyId" json:"accessKey,omitempty"`
	SecretKey    string    `xml:"SecretAccessKey" json:"secretKey,omitempty"`
	CreatedAt    time.Time `xml:"-" json:"createdAt,omitempty"`
	Expiration   time.Time `xml:"Expiration" json:"expiration,omitempty"`
	SessionToken string    `xml:"SessionToken" json:"sessionToken,omitempty"`
	Status       string    `xml:"-" json:"status,omitempty"`
}

// IsAccessKeyValid - validate access key for right length.
func IsAccessKeyValid(accessKey string) bool {
	return len(accessKey) >= accessKeyMinLen
}

// IsSecretKeyValid - validate secret key for right length.
func IsSecretKeyValid(secretKey string) bool {
	return len(secretKey) >= secretKeyMinLen
}

var timeSentinel = time.Unix(0, 0).UTC()

// IsExpired - returns whether AccessKey is expired or not.
func (acc AccessKey) IsExpired() bool {
	if acc.Expiration.IsZero() || acc.Expiration == timeSentinel {
		return false
	}

	return acc.Expiration.Before(time.Now().UTC())
}

// IsValid - returns whether AccessKey is valid or not.
func (acc AccessKey) IsValid() bool {
	// Verify credentials if its enabled or not set.
	if acc.Status == "enabled" || acc.Status == "" {
		return IsAccessKeyValid(acc.AccessKey) && IsSecretKeyValid(acc.SecretKey) && !acc.IsExpired()
	}
	return false
}

// Equal - returns whether two accesskeys are equal or not.
func (acc AccessKey) Equal(acc2 AccessKey) bool {
	if !acc2.IsValid() {
		return false
	}
	return (acc.AccessKey == acc2.AccessKey && subtle.ConstantTimeCompare([]byte(acc.SecretKey), []byte(acc2.SecretKey)) == 1 &&
		subtle.ConstantTimeCompare([]byte(acc.SessionToken), []byte(acc2.SessionToken)) == 1)
}

func expToInt64(expI interface{}) (expAt int64, err error) {
	switch exp := expI.(type) {
	case float64:
		expAt = int64(exp)
	case int64:
		expAt = exp
	case json.Number:
		expAt, err = exp.Int64()
		if err != nil {
			return 0, err
		}
	case time.Duration:
		return time.Now().UTC().Add(exp).Unix(), nil
	case nil:
		return 0, nil
	default:
		return 0, errors.New("invalid expiry value")
	}
	return expAt, nil
}

// GetNewAccessKeyWithMetadata generates and returns new credential with expiry.
func GetNewAccessKeyWithMetadata(m map[string]interface{}, tokenSecret string) (acc AccessKey, err error) {
	keyStr, err := text.RandomString(accessKeyMaxLen, AccessKeyChars)
	if err != nil {
		return
	}
	acc.AccessKey = keyStr

	// Generate secret key.
	keyStr, err = text.RandomString(secretKeyMaxLen, SecretKeyChars)
	if err != nil {
		return
	}
	acc.SecretKey = keyStr
	acc.Status = "enabled"

	expiry, err := expToInt64(m["exp"])
	if err != nil {
		return acc, err
	}
	if expiry == 0 {
		acc.Expiration = timeSentinel
		return acc, nil
	}

	m["accessKey"] = acc.AccessKey
	jwt := jwtgo.NewWithClaims(jwtgo.SigningMethodHS512, jwtgo.MapClaims(m))

	acc.Expiration = time.Unix(expiry, 0)
	acc.SessionToken, err = jwt.SignedString([]byte(tokenSecret))
	if err != nil {
		return acc, err
	}

	acc.CreatedAt = time.Now().UTC()

	return acc, nil
}

// GetNewCredentials generates and returns new credential.
func GetNewAccessKey() (acc AccessKey, err error) {
	return GetNewAccessKeyWithMetadata(map[string]interface{}{}, "")
}

// CreateCredentials returns new credential with the given access key and secret key.
// Error is returned if given access key or secret key are invalid length.
func CreateAccessKey(accessKey, secretKey string) (acc AccessKey, err error) {
	if !IsAccessKeyValid(accessKey) {
		return acc, ErrInvalidAccessKeyLength
	}
	if !IsSecretKeyValid(secretKey) {
		return acc, ErrInvalidSecretKeyLength
	}
	acc.AccessKey = accessKey
	acc.SecretKey = secretKey
	acc.Expiration = timeSentinel
	acc.CreatedAt  = time.Now().UTC()
	acc.Status = "enabled"
	return acc, nil
}
