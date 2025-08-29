package main

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
)

type redisConn struct {
	addr     string
	password string
}

func newRedisConn(addr string, password string) *redisConn {
	return &redisConn{
		addr:     addr,
		password: password,
	}
}

func (r *redisConn) createRedisInstance() (*redis.Client, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     r.addr,
		Password: r.password,
		DB:       0,
	})

	result, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	if result != "PONG" {
		return nil, errors.New("failed to connect to redis")
	}

	return rdb, nil
}
