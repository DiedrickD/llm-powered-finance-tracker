package initializers

import "github.com/DiedrickD/llm-powered-finance-tracker/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})
}