package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/DiedrickD/llm-powered-finance-tracker/initializers"
	"github.com/DiedrickD/llm-powered-finance-tracker/models"
	"github.com/DiedrickD/llm-powered-finance-tracker/services"
)

func CreateAutoTransaction(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get user from context in middleware requireAuth.go
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Description string `json:"description"`
		IsLlmActive bool   `json:"is_llm_active"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	if body.Description == "" {
		http.Error(w, "Description cannot be empty", http.StatusBadRequest)
		return
	}

	var aiCategory string
	var aiAmount int

	// If true then pass to LLM for auto categorization
	if body.IsLlmActive == true {
		// Get category list
		categoryList := GetCategories(user.ID)
		aiAmount, aiCategory, _ = services.GptService(r.Context(), body.Description, categoryList)

	}

	// Find or create category
	category, err := FindOrCreateCategory(aiCategory, user.ID)
	if err != nil {
		http.Error(w, "Failed to process category", http.StatusInternalServerError)
		return
	}

	// Create transaction
	transaction := models.Transaction{
		Amount:      aiAmount,
		UserID:      user.ID,
		Description: body.Description,
		Categories:  []models.Category{category},
	}

	// Put transaction into database
	result := initializers.DB.Create(&transaction)
	if result.Error != nil {
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}

	// Respond
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

// Usage GET /transactions?start_date=01/01/2025&end_date=31/01/2025
func ReadTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var transactions []models.Transaction

	// Ambil query param
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	db := initializers.DB.Where("user_id = ?", user.ID)

	// If the date param active use filter
	if startDateStr != "" && endDateStr != "" {
		layout := "02/01/2006" 

		startDate, err := time.Parse(layout, startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format (dd/mm/yyyy)", http.StatusBadRequest)
			return
		}

		endDate, err := time.Parse(layout, endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format (dd/mm/yyyy)", http.StatusBadRequest)
			return
		}

		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

		db = db.Where("created_at BETWEEN ? AND ?", startDate, endDate)
	}

	// Query data
	result := db.
		Order("created_at DESC").
		Find(&transactions)

	if result.Error != nil {
		http.Error(w, "Failed to retrieve transactions", http.StatusInternalServerError)
		return
	}

	// Return to user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}


// Usage: https/link/updateTransaction
func UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get value from user
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		TransactionID int      `json:"transaction_id"`
		Description   *string  `json:"description"`
		Amount        *int     `json:"amount"`
		Categories    []string `json:"categories"`
		IsLlmActive   bool     `json:"is_llm_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	var transaction models.Transaction
	if err := initializers.DB.
		Preload("Categories").
		Where("id = ? AND user_id = ?", body.TransactionID, user.ID).
		First(&transaction).Error; err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	// Auto update using LLM
	if body.IsLlmActive && body.Description != nil {
		categoryList := GetCategories(user.ID)
		aiAmount, aiCategory, err := services.GptService(
			r.Context(),
			*body.Description,
			categoryList,
		)

		if err == nil {
			transaction.Amount = aiAmount

			category, err := FindOrCreateCategory(aiCategory, user.ID)
			if err == nil {
				initializers.DB.Model(&transaction).
					Association("Categories").
					Replace([]models.Category{category})
			}
		}
	}

	if body.Amount != nil {
		transaction.Amount = *body.Amount
	}

	if body.Description != nil {
		transaction.Description = *body.Description
	}

	if len(body.Categories) > 0 {
		var updatedCategories []models.Category

		for _, name := range body.Categories {
			category, err := FindOrCreateCategory(name, user.ID)
			if err == nil {
				updatedCategories = append(updatedCategories, category)
			}
		}

		initializers.DB.Model(&transaction).
			Association("Categories").
			Replace(updatedCategories)
	}

	if err := initializers.DB.Save(&transaction).Error; err != nil {
		http.Error(w, "Failed to update transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(transaction)
}


func DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	// Check for method
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get the parameter for transaction ID from query
	q := r.URL.Query()

	idInt, err := strconv.Atoi(q.Get("id"))
	if err != nil {
		http.Error(w, "ID must be a number", http.StatusBadRequest)
		return
	}

	idUint := uint(idInt)

	// Get user info from middleware
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unathorized", http.StatusUnauthorized)
		return
	}

	var transaction models.Transaction

	// Delete the data on the database
	result := initializers.DB.Delete(&transaction, "id = ? AND user_id = ?", idUint, user.ID)
	if result.RowsAffected == 0 {
		http.Error(w, "Transaction not found or you don't have permission", http.StatusNotFound)
		return
	}

	// Return in json
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Data deleted",
	})
}
