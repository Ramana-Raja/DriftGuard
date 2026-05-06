package tasks

import (
	"DriftGuard/backend/database"
	"log"
)

func AllActive() {
	var results []database.ModelVersion

	err := database.DB.
		Where("is_active = ?", true).
		Find(&results).Error

	if err != nil {
		log.Println("failed to fetch active models:", err)
		return
	}

	for _, r := range results {
		err := DispatchDriftTask(
			r.FilePath,
			r.DatasetLink,
		)
		if err != nil {
			log.Println("failed to dispatch task:", err)
		}
	}
}
