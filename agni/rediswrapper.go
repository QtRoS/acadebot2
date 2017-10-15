package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/QtRoS/acadebot2/shared/logu"
	"github.com/go-redis/redis"
)

var client *redis.Client

const (
	envRedisAddress   = "ENV_REDIS_ADDRESS"
	envRedisPass      = "ENV_REDIS_PASS"
	searchTTLMinutes  = 60
	contextTTLMinutes = 15
)

type userContext struct {
	Query    string `json:"query"`
	Count    int    `json:"count"`
	Position int    `json:"position"`
}

func init() {
	client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv(envRedisAddress) + ":6379",
		Password: os.Getenv(envRedisPass), // no password set
		DB:       0,                       // use default DB
	})

	pong, err := client.Ping().Result()
	logu.Info.Println("Redis ping: ", pong, err)
}

// saveContext to Redis.
func saveContext(chatID int64, context *userContext) {
	redisKey := fmt.Sprintf("usercontext:%d", chatID)
	value, err := json.Marshal(context)
	if err != nil {
		logu.Error.Println(err)
		return
	}

	client.Set(redisKey, value, time.Minute*contextTTLMinutes)
}

// restoreContext from Redis.
func restoreContext(chatID int64) *userContext {
	redisKey := fmt.Sprintf("usercontext:%d", chatID)

	value, err0 := client.Get(redisKey).Result()
	if err0 != nil {
		if err0 == redis.Nil {
			logu.Warning.Println("No context for:", redisKey)
		} else {
			logu.Error.Println("Redis error:", err0)
		}
		return nil
	}

	var context userContext
	err1 := json.Unmarshal([]byte(value), &context)
	if err1 != nil {
		logu.Error.Println(err1)
	}
	return &context
}
