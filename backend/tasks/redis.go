package tasks

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var RedisClient *redis.Client

const (
	StreamName = "drift_check_stream"
	GroupName  = "drift_workers"
)

type DriftTask struct {
	StoragePath string `json:"storage_path"`
	DatasetLink string `json:"dataset_link"`
	DatasetPath string `json:"dataset_path"`
	ProjectId   int    `json:"project_id"`
}

func RedisInt() {

	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     host + ":6379",
		Password: "",
		DB:       0,
	})
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("could not connect to Redis: %v", err)
	}

	fmt.Println("connected to Redis")

	// Ensure the stream and consumer group exist for the ACK-based architecture
	err = RedisClient.XGroupCreateMkStream(ctx, StreamName, GroupName, "$").Err()
	if err != nil {
		// Ignore if the consumer group already exists (BUSYGROUP)
		if strings.Contains(err.Error(), "BUSYGROUP") {
			log.Printf("Redis consumer group '%s' already initialized on stream '%s'", GroupName, StreamName)
		} else {
			log.Printf("Warning: failed to initialize consumer group: %v", err)
		}
	} else {
		log.Printf("Initialized Redis stream '%s' and consumer group '%s'", StreamName, GroupName)
	}
}

func DispatchDriftTask(storagePath string, datasetlink string, datasetpath string, projectid int) error {
	taskValues := map[string]interface{}{
		"storage_path": storagePath,
		"dataset_link": datasetlink,
		"dataset_path": datasetpath,
		"project_id":   projectid,
	}

	// XAdd appends to the Redis Stream.
	// The message remains persistent and moves to the Pending Entries List (PEL)
	// when a consumer reads it, remaining there until explicitly acknowledged via XACK.
	msgID, err := RedisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: StreamName,
		Values: taskValues,
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to append task to stream %s: %w", StreamName, err)
	}

	fmt.Printf("dispatched task [ID: %s] to stream '%s' for project %d (path: %s)\n", msgID, StreamName, projectid, storagePath)
	return nil
}
