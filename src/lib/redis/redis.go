package redis

import (
	redigo "github.com/gomodule/redigo/redis"
	"time"
)

var RedisPool *redigo.Pool

func newPool(server, password string, maxIdle, maxActive, idelTimeout int) *redigo.Pool {
	return &redigo.Pool {
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: time.Duration(idelTimeout) * time.Second,
		Dial: func() (redigo.Conn, error) {
			c, err := redigo.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err = c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}

			_, err := c.Do("PING")
			return err
		},
	}
}

func InitRedisPool(server, password string, maxIdle, maxActive, idleTimeout int) {
	RedisPool = newPool(server, password, maxIdle, maxActive, idleTimeout)
	return
}
