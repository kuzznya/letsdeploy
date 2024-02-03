package redisclient

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func New(cfg *viper.Viper) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: cfg.GetString("redisclient.host"),
	})
}
