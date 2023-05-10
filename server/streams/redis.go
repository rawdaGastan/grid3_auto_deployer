// Package streams for API endpoints
package streams

import (
	"github.com/codescalers/cloud4students/internal"
	"github.com/go-redis/redis"
)

// RedisClient for redis DB handling streams
type RedisClient struct {
	DB *redis.Client
}

// NewRedisClient creates a new RedisClient
func NewRedisClient(config internal.Configuration) (RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Server.RedisHost + ":" + config.Server.RedisPort,
		Password: config.Server.RedisPass,
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return RedisClient{}, err
	}

	client.XGroupCreateMkStream(DeployK8sStreamName, DeployK8sConsumerGroupName, "$")
	client.XGroupCreateMkStream(DeployVMStreamName, DeployVMConsumerGroupName, "$")
	client.XGroupCreateMkStream(DeployNetStreamName, DeployNetConsumerGroupName, "$")
	client.XGroupCreateMkStream(ReqVMStreamName, ReqVMConsumerGroupName, "$")
	client.XGroupCreateMkStream(ReqK8sStreamName, ReqK8sConsumerGroupName, "$")

	return RedisClient{client}, nil
}
