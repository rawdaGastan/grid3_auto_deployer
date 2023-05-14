// Package streams for redis streams
package streams

import (
	"time"

	"github.com/go-redis/redis"
)

func (r *RedisClient) Read(stream, group string, pending bool) (result []redis.XStream, err error) {
	IDs := ">"
	if pending {
		IDs = "0"
	}
	result, err = r.DB.XReadGroup(&redis.XReadGroupArgs{
		Streams: []string{stream, IDs},
		Group:   group,
		Block:   1 * time.Second,
	}).Result()

	return
}
