package weblimiter

import (
	"encoding/json"

	redis "github.com/go-redis/redis/v7"
)

type RedisClient struct {
	client *redis.Client
	parser func(rules map[string]string) ([]RateConf, error)
}

func NewRedisClient(options *redis.Options, parser func(rules map[string]string) ([]RateConf, error)) *RedisClient {
	client := redis.NewClient(options)
	return &RedisClient{
		client: client,
		parser: parser,
	}
}

func (client *RedisClient) GetConfig(key string) (map[string]string, error) {
	value, err := client.client.Get(key).Result()
	if err != nil {
		return nil, err
	}

	dump := make(map[string]string)
	if err := json.Unmarshal([]byte(value), &dump); err != nil {
		return dump, err
	}
	return dump, nil
}

func (client *RedisClient) ParseConfig(config map[string]string) ([]RateConf, error) {
	return client.parser(config)
}
