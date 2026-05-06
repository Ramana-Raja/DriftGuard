package database

import "gorm.io/gorm"

type UserAuth struct {
	gorm.Model
	Email    string `json:"email" gorm:"unique;not null"`
	Password string `json:"-"`

	Profile UserProfile `json:"profile"`
}

type UserProfile struct {
	gorm.Model
	UserAuthID  uint   `json:"-"`
	DisplayName string `json:"display_name"`

	Projects []Project `json:"projects"`
}
type Project struct {
	gorm.Model
	UserProfileID uint   `json:"-"`
	ProjectName   string `json:"project_name"`
	Description   string `json:"description"`

	ProjectModel Model `gorm:"foreignKey:ProjectID"`
}

type Model struct {
	gorm.Model
	ProjectID uint

	Name string

	Versions []ModelVersion
}

type ModelVersion struct {
	gorm.Model
	ModelID     uint
	DatasetLink string //holds the latest dataset from the source
	Version     int
	FilePath    string
	IsActive    bool
	DatasetPath string //holds the current dataset
}
