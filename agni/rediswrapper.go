package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/QtRoS/acadebot2/shared/logu"
	"github.com/QtRoS/acadebot2/shared/netu"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

var client *redis.Client

const (
	SearchUrl         = "http://localhost:19191/courses"
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
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
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

func Search(query string, limit int) string {
	redisKey := fmt.Sprintf("query:%x", md5.Sum([]byte(query)))

	value, err := client.Get(redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			logu.Info.Println("Redis miss: ", query)
			newValue := searchInBackService(query, limit)
			client.Set(redisKey, newValue, time.Minute*SearchTtlMinutes)
			value = string(newValue)
		} else {
			logu.Error.Println("Redis error:", err)
		}
	}

	return value
}

func searchInBackService(query string, limit int) []byte {
	data, err := netu.MakeRequest(SearchUrl,
		map[string]string{"query": query, "limit": strconv.Itoa(limit)}, nil)

	if err != nil {
		logu.Error.Println(err)
		return nil
	}

	return data
}
