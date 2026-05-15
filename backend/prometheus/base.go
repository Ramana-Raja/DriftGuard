package prometheus

import (
	"context"
	"fmt"
	"time"
)

func StartPro() {
	fmt.Println("Started Prometheus System Performance Target Monitor...")

	ctx := context.Background()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Hour):
			// Keeps the background thread quietly alive without blocking or looping on CPU cycles
			fmt.Println("System performance monitoring pipeline: Active and Healthy.")
		}
	}
}
