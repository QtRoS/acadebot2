package searchengine

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/QtRoS/acadebot2/shared"
	"github.com/QtRoS/acadebot2/shared/logu"
	"github.com/go-redis/redis"
)

const (
	envRedisAddress = "ENV_REDIS_ADDRESS"
	envRedisPass    = "ENV_REDIS_PASS"
)

type cachingAdapter struct {
	sourceAdapter SourceAdapter
	client        *redis.Client
	ttl           time.Duration
}

func newCachingAdapter(adapter SourceAdapter, ttl time.Duration) *cachingAdapter {
	ca := cachingAdapter{}
	ca.sourceAdapter = adapter
	ca.ttl = ttl

	ca.client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv(envRedisAddress) + ":6379",
		Password: os.Getenv(envRedisPass), // no password set
		DB:       0,                       // use default DB
	})

	pong, err := ca.client.Ping().Result()
	logu.Info.Println("Caching client", ca.sourceAdapter.Name(), "redis ping: ", pong, err)

	return &ca
}

func (me *cachingAdapter) Name() string {
	return me.sourceAdapter.Name() + " (Cached)"
}

func (me *cachingAdapter) Get(query string, limit int) []shared.CourseInfo {
	redisKey := fmt.Sprintf("cachingadapter:%s", me.sourceAdapter.Name())

	value, err := me.client.Get(redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			logu.Info.Println("Redis miss:", redisKey)
		} else {
			logu.Error.Println("Redis error:", err)
		}

		rawData := me.sourceAdapter.Get(query, limit)
		rawDataAsJSON, err := json.Marshal(rawData)
		if err != nil {
			logu.Error.Println(err)
		} else if rawData != nil {
			me.client.Set(redisKey, rawDataAsJSON, me.ttl)
		}

		return rawData
	}

	logu.Error.Println("Redis HIT:", redisKey)
	var result []shared.CourseInfo
	err1 := json.Unmarshal([]byte(value), &result)
	if err1 != nil {
		logu.Error.Println(err1)
	}
	return result
}
