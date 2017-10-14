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
	EnvRedisAddress   = "ENV_REDIS_ADDRESS"
	EnvRedisPass      = "ENV_REDIS_PASS"
	SearchTtlMinutes  = 60
	ContextTtlMinutes = 15
)

type UserContext struct {
	Query    string `json:"query"`
	Count    int    `json:"count"`
	Position int    `json:"position"`
}

func init() {
	client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv(EnvRedisAddress) + ":6379",
		Password: os.Getenv(EnvRedisPass), // no password set
		DB:       0,                       // use default DB
	})

	pong, err := client.Ping().Result()
	logu.Info.Println("Redis ping: ", pong, err)
}

func SaveContext(chatId int64, context *UserContext) {
	redisKey := fmt.Sprintf("usercontext:%d", chatId)
	value, err := json.Marshal(context)
	if err != nil {
		logu.Error.Println(err)
		return
	}

	client.Set(redisKey, value, time.Minute*ContextTtlMinutes)
}

func RestoreContext(chatId int64) *UserContext {
	redisKey := fmt.Sprintf("usercontext:%d", chatId)

	value, err0 := client.Get(redisKey).Result()
	if err0 != nil {
		if err0 == redis.Nil {
			logu.Warning.Println("No context for:", redisKey)
		} else {
			logu.Error.Println("Redis error:", err0)
		}
		return nil
	}

	var context UserContext
	err1 := json.Unmarshal([]byte(value), &context)
	if err1 != nil {
		logu.Error.Println(err1)
	}
	return &context
}
