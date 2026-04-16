package database

import "gorm.io/gorm"

type UserAuth struct {
	gorm.Model
	Email       string `json:"email" gorm:"unique;not null"`
	Password    string `json:"-"`
	LastLoginIP string `json:"last_login_ip"`

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

	Models []Model `json:"models"`
}

type Model struct {
	gorm.Model
	ProjectID   uint   `json:"project_id"`
	Name        string `json:"name"`
	StoragePath string `json:"storage_path"`
	Framework   string `json:"framework"`
}
