/**
 * @Author: Administrator
 * @Description:
 * @File: init
 * @Date: 2022/4/29 11:16
 */
package DB

import (
	"github.com/go-redis/redis"
)

var RDB *RedisCache

type RedisCache struct {
	Client *redis.Client
}

func init() {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	RDB = &RedisCache{
		Client: client,
	}
}

func (r *RedisCache) Set(key string, value interface{}) error {
	return r.Client.Set(key, value, 0).Err()
}

func (r *RedisCache) Get(key string) (string, bool, error) {
	s, err := r.Client.Get(key).Result()
	if err == redis.Nil {
		return "", false, err
	} else if err != nil {
		return "", true, err
	}
	return s, true, err
}

func (r *RedisCache) Del(key string) (bool, error) {
	_, err := r.Client.Del(key).Result()
	if err == redis.Nil {
		return false, err
	} else if err != nil {
		return true, err
	}
	return true, err
}
