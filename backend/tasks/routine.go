package tasks

import (
	"log"
	"time"
)

func StartDriftCheck() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		log.Println("Running drift check...")
		AllActive()

		<-ticker.C
	}
}
