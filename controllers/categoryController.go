package controllers

import (
	"errors"
	"fmt"

	"github.com/DiedrickD/llm-powered-finance-tracker/initializers"
	"github.com/DiedrickD/llm-powered-finance-tracker/models"
	"gorm.io/gorm"
)

func GetCategories(userID uint) []string {
	var categories []models.Category
	var listCategory []string

	// Get category from user and maincategory (0)
	result := initializers.DB.Limit(10).Where("user_id IN ?", []uint{0, userID}).Find(&categories)

	if result.Error != nil {
		return []string{}
	}

	for _, val := range categories {
		listCategory = append(listCategory, val.Name)
	}

	fmt.Println(listCategory)
	return listCategory
}

func FindOrCreateCategory(name string, userID uint) (models.Category, error) {
	var category models.Category

	// Find user category 
	err := initializers.DB.Where("name = ? AND user_id = ?", name, userID).First(&category).Error
	if err == nil {
		return category, nil
	}
	
	// Find main category 
	err = initializers.DB.Where("name = ? AND user_id = ?", name, 0).First(&category).Error
	if err == nil {
		return category, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		category = models.Category{
			Name:   name,
			UserID: userID,
		}
		err = initializers.DB.Create(&category).Error
	}

	return category, err
}

//func deleteCategories() {

//}
