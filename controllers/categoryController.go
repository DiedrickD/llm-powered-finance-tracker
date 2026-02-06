package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

func DeleteCategoryByID(categoryID uint, userID uint) error {
	var category models.Category

	err := initializers.DB.
		Where("id = ? AND user_id = ?", categoryID, userID).
		First(&category).Error

	if err != nil {
		return err
	}

	return initializers.DB.Delete(&category).Error
}

func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(uint)

	categoryIDStr := r.URL.Query().Get("id")
	if categoryIDStr == "" {
		http.Error(w, "category id is required", http.StatusBadRequest)
		return
	}

	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid category id", http.StatusBadRequest)
		return
	}

	err = DeleteCategoryByID(uint(categoryID), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
