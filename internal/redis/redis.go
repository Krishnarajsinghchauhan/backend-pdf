package redis

import (
	"context"
	"os"

	redis "github.com/redis/go-redis/v9"
	"encoding/json"
)

var ctx = context.Background()
var client *redis.Client

func Init() {
	client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

func SaveJob(jobID, status string) {
	client.Set(ctx, "job:"+jobID, status, 0)
}

func GetJobStatus(jobID string) string {
	val, err := client.Get(ctx, "job:"+jobID).Result()
	if err != nil {
		return "not_found"
	}
	return val
}

func SaveResult(jobID string, urls []string) {
	b, _ := json.Marshal(urls)
	client.Set(ctx, "result:"+jobID, b, 0)
	client.Set(ctx, "job:"+jobID, "completed", 0)
}

func GetResult(jobID string) string {
	val, _ := client.Get(ctx, "result:"+jobID).Result()
	return val
}

