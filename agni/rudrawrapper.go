package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/QtRoS/acadebot2/shared/logu"
	"github.com/QtRoS/acadebot2/shared/netu"
	"github.com/go-redis/redis"
)

const EnvRudraAddress = "ENV_RUDRA_ADDRESS"

var searchURL = "http://" + os.Getenv(EnvRudraAddress) + ":19191/courses"

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
	data, err := netu.MakeRequest(searchURL,
		map[string]string{"query": query, "limit": strconv.Itoa(limit)}, nil)

	if err != nil {
		logu.Error.Println(err)
		return nil
	}

	return data
}