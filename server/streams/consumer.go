// Package streams for redis streams
package streams

import (
	"errors"
	"time"

	"github.com/go-redis/redis"
)

func (r *RedisClient) Read(stream, group string) (result []redis.XStream, err error) {
	result, err = r.DB.XReadGroup(&redis.XReadGroupArgs{
		Streams: []string{stream, ">"},
		Group:   group,
		Block:   1 * time.Second,
	}).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		return
	}

	pending, err := r.DB.XReadGroup(&redis.XReadGroupArgs{
		Streams: []string{stream, "0"},
		Group:   group,
		Block:   1 * time.Second,
	}).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		return
	}

	result = append(result, pending...)
	if len(result) == 0 {
		return result, redis.Nil
	}

	return
}
