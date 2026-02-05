package initializers

import "github.com/DiedrickD/llm-powered-finance-tracker/models"

func SeedCategories() {
	// Seed sub categorynya juga
	mainCategory := []string{"Income", "Expense", "Transfer", "Debt"}

	for _, name := range mainCategory {
		var category models.Category
		DB.FirstOrCreate(&category, models.Category{
			Name:   name,
			UserID: 0,
		})
	}
}