package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var RedisClient *redis.Client

type DriftTask struct {
	StoragePath string `json:"storage_path"`
	DatasetLink string `json:"dataset_link"`
}

func RedisInt() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("could not connect to Redis: %v", err)
	}
	fmt.Println("connected to Redis")
}

func DispatchDriftTask(storagePath string, datasetlink string) error {
	task := DriftTask{
		StoragePath: storagePath,
		DatasetLink: datasetlink,
	}

	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = RedisClient.LPush(ctx, "drift_check_queue", taskJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to push task: %w", err)
	}
	fmt.Printf("dispatched task for path: %s\n", storagePath)
	return nil
}
