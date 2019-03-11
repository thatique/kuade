package redis

import (
	"crypto/subtle"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/thatique/kuade/kuade/auth"
	"github.com/thatique/kuade/kuade/auth/passwords/tokens"
)

type RedisToken struct {
	pool      *redis.Pool
	Expire    int
	keyPrefix string
}

func NewRedisToken(pool *redis.Pool) *RedisToken {
	return &RedisToken{
		pool: pool,
		Expire: 7200, // two hours
		keyPrefix: "kuade:tokens:",
	}
}

func (t *RedisToken) SetKeyPrefix(prefix string) {
	t.keyPrefix = prefix
}

var insertScript = redis.NewScript(2, `
	local tokens = {}
	for i = 2, #ARGV, 1 do
		tokens[#tokens + 1] = ARGV[i]
	end

	redis.call('HMSET', KEYS[1], unpack(tokens))
	redis.call('SET', KEYS[2], KEYS[1])

	if(ARGV[1] ~= '') then
		redis.call('EXPIRE', KEYS[1], ARGV[1])
	end

	return true
`)

func (t *RedisToken) Generate(user *auth.User) (token string, err error) {
	conn := t.pool.Get()
	defer conn.Close()

	if err = conn.Err(); err != nil {
		return "", err
	}

	token, err = tokens.GenerateToken()
	if err != nil {
		return "", err
	}

	tok := &tokens.PasswordToken{
		Token:     token,
		Email:     user.Email,
		Pass:      user.Password,
		CreatedAt: time.Now().UTC().Unix(),
	}

	args := redis.Args{}.Add(t.keyPrefix+user.Id.Hex()).Add(t.keyPrefix+token)
	args = args.Add(t.Expire).AddFlat(tok)
	_, err = insertScript.Do(conn, args...)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (t *RedisToken) Delete(token string) error {
	conn := t.pool.Get()
	defer conn.Close()

	if err := conn.Err(); err != nil {
		return err
	}

	tok := t.keyPrefix+token
	key, err := redis.String(conn.Do("GET", tok))
	if err != nil {
		return err
	}

	conn.Send("MULTI")
	conn.Send("DEL", tok)
	conn.Send("DEL", key)
	_, err = conn.Do("EXEC")

	return err
}

func (t *RedisToken) IsValid(user *auth.User, token string) bool {
	conn := t.pool.Get()
	defer conn.Close()

	if err := conn.Err(); err != nil {
		return false
	}

	data, err := redis.Values(conn.Do("HGETALL", t.keyPrefix+user.Id.Hex()))
	if err != nil {
		return false
	}

	if len(data) == 0 {
		return false
	}

	var stored = new(tokens.PasswordToken)
	if err = redis.ScanStruct(data, stored); err != nil {
		return false
	}

	if stored.Email != user.Email || stored.Token != token || subtle.ConstantTimeCompare(stored.Pass, user.Password) != 1 {
		return false
	}

	return true
}