package initializers

import "github.com/DiedrickD/llm-powered-finance-tracker/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.Transaction{})
	DB.AutoMigrate(&models.Category{})
}