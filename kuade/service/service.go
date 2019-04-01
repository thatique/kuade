package service

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/thatique/kuade/configuration"
	"github.com/thatique/kuade/kuade/storage"
	"github.com/thatique/kuade/kuade/storage/factory"
	"github.com/thatique/kuade/pkg/mailer"
	_ "github.com/thatique/kuade/pkg/mailer/smtp"
	"github.com/thatique/kuade/pkg/queue"
)

type Service struct {
	Storage storage.Driver
	Queue   *queue.Queue
	Redis   *redis.Pool
	Mailer  *mailer.Transport
	Config  *configuration.Configuration
}

func NewService(conf *configuration.Configuration) (*Service, error) {
	storage, err := factory.Create(conf.Storage.Type(), conf.Storage.Parameters())
	if err != nil {
		return nil, err
	}
	var redisPool *redis.Pool
	if conf.Redis.Addr != "" {
		redisPool, err = newRedisPool(conf.Redis)
	}
	mailer := setupSMTPTransport(conf.Mail)
	q := configureQueue(conf.Queue)

	return &Service{
		Storage: storage,
		Queue:   q,
		Redis:   redisPool,
		Mailer:  mailer,
		Config:  conf,
	}, nil
}

func (service *Service) Quit() {
	service.Queue.Stop()
}

func configureQueue(conf configuration.Queue) *queue.Queue {
	if conf.MaxWorkers == 0 {
		conf.MaxWorkers = 5
	}
	if conf.MaxQueue == 0 {
		conf.MaxQueue = 100
	}
	return queue.NewQueue(conf.MaxWorkers, conf.MaxQueue)
}

func dialRedisWithConf(conf configuration.Redis) (redis.Conn, error) {
	conn, err := redis.DialTimeout("tcp",
		conf.Addr,
		conf.DialTimeout,
		conf.ReadTimeout,
		conf.WriteTimeout)

	if err != nil {
		return nil, err
	}

	if conf.Password != "" {
		// do auth
		if _, err := conn.Do("AUTH", conf.Password); err != nil {
			conn.Close()
			return nil, err
		}
	}

	// select DB if asked
	if conf.DB != 0 {
		if _, err = conn.Do("SELECT", conf.DB); err != nil {
			conn.Close()
			return nil, err
		}
	}

	return conn, nil
}

// Create redis Pool
func newRedisPool(conf configuration.Redis) (*redis.Pool, error) {
	pool := &redis.Pool{
		MaxIdle:     conf.MaxIdle,
		MaxActive:   conf.MaxActive,
		IdleTimeout: conf.IdleTimeout,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			return dialRedisWithConf(conf)
		},
	}

	// test the connection
	_, err := pingRedis(pool)
	return pool, err
}

// Ping against a server to check if it is alive.
func pingRedis(pool *redis.Pool) (bool, error) {
	conn := pool.Get()
	defer conn.Close()
	data, err := conn.Do("PING")
	return (data == "PONG"), err
}

func setupSMTPTransport(conf configuration.Mail) *mailer.Transport {
	t, err := mailer.OpenTransport(context.Background(), conf.TransportURL)
	if err != nil {
		panic(err)
	}
	return t
}
