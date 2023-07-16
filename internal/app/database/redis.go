package database

import redis "github.com/go-redis/redis/v7"

var (
	redisClient *redis.Client
)

func InitRedis(dsn string) {
	redisClient = redis.NewClient(&redis.Options{
		Addr: dsn,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
}

func GetRedisClient(dsn string) *redis.Client {
	if redisClient == nil {
		InitRedis(dsn)
	}
	return redisClient
}
