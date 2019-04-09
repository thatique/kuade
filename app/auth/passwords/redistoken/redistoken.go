package redistoken

import (
	"crypto/subtle"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/thatique/kuade/app/auth/passwords"
	"github.com/thatique/kuade/app/model"
)

// Lua script for inserting password change token
//
// KEYS[1] - tokenid
// KEYS[2] - token
// ARGV[1] - Expiration in seconds
// ARGV[2] - The password token in bytes
var insertScript = redis.NewScript(2, `
	redis.call('SET', KEYS[1], ARGV[2])
	redis.call('SET', KEYS[2], KEYS[1])
	redis.call('EXPIRE', KEYS[1], ARGV[1])
	redis.call('EXPIRE', KEYS[2], ARGV[1])

	return true
`)

type RedisTokenGenerator struct {
	pool      *redis.Pool
	Expire    int
	keyPrefix string
}

func New(pool *redis.Pool) *RedisTokenGenerator {
	return &RedisTokenGenerator{
		pool:      pool,
		Expire:    7200,
		keyPrefix: "kuade:passwordtoken:",
	}
}

func (gen *RedisTokenGenerator) SetKeyPrefix(s string) {
	gen.keyPrefix = s
}

func (gen *RedisTokenGenerator) Generate(user *model.User) (token string, err error) {
	conn := gen.pool.Get()
	defer conn.Close()

	if err := conn.Err(); err != nil {
		return "", err
	}

	token, err = passwords.GenerateToken()
	if err != nil {
		return "", nil
	}

	pswdToken := &model.PasswordChangeToken{
		Token:     token,
		Email:     user.Email,
		Password:  user.Credentials.Password,
		CreatedAt: time.Now().UTC().Unix(),
	}
	data, err := pswdToken.Marshal()
	if err != nil {
		return "", err
	}

	_, err = insertScript.Do(conn, gen.keyPrefix+user.ID.Hex(), gen.keyPrefix+token, gen.Expire, data)
	return token, err
}

func (gen *RedisTokenGenerator) Delete(token string) error {
	conn := gen.pool.Get()
	defer conn.Close()

	if err := conn.Err(); err != nil {
		return err
	}

	tokKey := gen.keyPrefix + token
	key, err := redis.String(conn.Do("GET", tokKey))
	if err != nil {
		return err
	}

	_, err = conn.Do("DEL", tokKey, key)

	return err
}

func (gen *RedisTokenGenerator) IsValid(user *model.User, token string) bool {
	conn := gen.pool.Get()
	defer conn.Close()

	if err := conn.Err(); err != nil {
		return false
	}

	data, err := redis.Bytes(conn.Do("GET", gen.keyPrefix+user.ID.Hex()))
	if err != nil {
		return false
	}

	pswdToken := &model.PasswordChangeToken{}
	err = pswdToken.Unmarshal(data)
	if err != nil {
		return false
	}

	if pswdToken.Email != user.Email || pswdToken.Token != token {
		return false
	}

	expireAt := time.Unix(pswdToken.CreatedAt, 0).Add(time.Duration(gen.Expire) * time.Second)
	if time.Now().After(expireAt) {
		return false
	}

	return subtle.ConstantTimeCompare(pswdToken.Password, user.Credentials.Password) == 1
}
