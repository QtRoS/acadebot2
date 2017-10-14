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
	EmptyResult     = "[]"
	EnvRedisAddress = "ENV_REDIS_ADDRESS"
	EnvRedisPass    = "ENV_REDIS_PASS"
)

type cachingAdapter struct {
	sourceAdapter SourceAdapter
	client        *redis.Client
	ttl           time.Duration
}

func NewCachingAdapter(adapter SourceAdapter, ttl time.Duration) *cachingAdapter {
	ca := cachingAdapter{}
	ca.sourceAdapter = adapter
	ca.ttl = ttl

	ca.client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv(EnvRedisAddress) + ":6379",
		Password: os.Getenv(EnvRedisPass), // no password set
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
	//redisKey := fmt.Sprintf("cachingadapter:%s:%x", me.sourceAdapter.Name(), md5.Sum([]byte(query)))
	redisKey := fmt.Sprintf("cachingadapter:%s", me.sourceAdapter.Name())

	value, err := me.client.Get(redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			logu.Info.Println("cachingAdapter redis miss:", redisKey)
			newValue := me.sourceAdapter.Get(query, limit)
			valueAsJSON, err := json.Marshal(newValue)
			if err != nil {
				logu.Error.Println(err)
			} else if newValue != nil {
				me.client.Set(redisKey, valueAsJSON, me.ttl)
			}

			return newValue
		}

		logu.Error.Println("cachingAdapter redis error:", err)
		return nil
	}

	logu.Error.Println("cachingAdapter hit:", redisKey)
	var result []shared.CourseInfo
	err1 := json.Unmarshal([]byte(value), &result)
	if err1 != nil {
		logu.Error.Println(err1)
	}
	return result
}
