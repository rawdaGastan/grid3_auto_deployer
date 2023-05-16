// Package streams for redis streams
package streams

import (
	"time"

	"github.com/go-redis/redis"
)

func (r *RedisClient) Read(stream, group string, count int64, pending bool) (result []redis.XStream, err error) {
	IDs := ">"
	if pending {
		IDs = "0"
	}

	args := redis.XReadGroupArgs{
		Streams: []string{stream, IDs},
		Group:   group,
		Block:   1 * time.Second,
	}

	if count != 0 {
		args.Count = count
	}

	result, err = r.DB.XReadGroup(&args).Result()
	return
}
