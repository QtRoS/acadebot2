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

const envRudraAddress = "ENV_RUDRA_ADDRESS"

var searchURL = "http://" + os.Getenv(envRudraAddress) + ":19191/courses"

func init() {
	netu.CommonClient.Timeout = netu.CommonClient.Timeout + 2*time.Second
}

// RudraSearch for courses in Rudra.
func RudraSearch(query string, limit int) string {
	redisKey := fmt.Sprintf("query:%x", md5.Sum([]byte(query)))

	value, err := client.Get(redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			logu.Info.Println("Redis miss: ", query)
		} else {
			logu.Error.Println("Redis error:", err)
		}
		newValue := rudraSearchInternal(query, limit)
		if newValue != nil {
			client.Set(redisKey, newValue, time.Minute*searchTTLMinutes)
		}
		value = string(newValue)
	}

	return value
}

func rudraSearchInternal(query string, limit int) []byte {
	data, err := netu.MakeRequest(searchURL,
		map[string]string{"query": query, "limit": strconv.Itoa(limit)}, nil)

	if err != nil {
		logu.Error.Println(err)
		return nil
	}

	return data
}
