package tasks

import (
	"DriftGuard/backend/database"
	"log"
)

func AllActive() {
	var results []string

	err := database.DB.
		Model(&database.ModelVersion{}).Where("is_active = ?", true).Pluck("file_path", &results).Error

	if err != nil {
		log.Println("failed to fetch active models:", err)
		return
	}

	for _, r := range results {
		err := DispatchDriftTask(
			r,
		)
		if err != nil {
			log.Println("failed to dispatch task:", err)
		}
	}
}
