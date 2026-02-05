package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/DiedrickD/llm-powered-finance-tracker/initializers"
	"github.com/DiedrickD/llm-powered-finance-tracker/models"
	"github.com/DiedrickD/llm-powered-finance-tracker/services"
)

func CreateTransaction(w http.ResponseWriter, r *http.Request) {
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
		Amount      int    `json:"amount"`
		Category    string `json:"category"`
		Description string `json:"description"`
		IsLlmActive bool   `json:"is_llm_active"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	// Error for amount 0
	if body.Amount == 0 {
		http.Error(w, "Amount cannot be 0", http.StatusBadRequest)
		return
	}

	if body.Description == "" {
		http.Error(w, "Description cannot be empty", http.StatusBadRequest)
		return
	}

	// If true then pass to LLM for auto categorization
	if body.IsLlmActive == true {
		// Get category list
		categoryList := GetCategories(user.ID)

		aiCategory, err := services.GptService(r.Context(), body.Description, categoryList)

		// If error happens, use Uncategorized, else use the ai category
		if err != nil {
			body.Category = "Uncategorized"
		} else {
			body.Category = aiCategory
		}
	}

	// If category is empty, categorize it as Uncategorized
	if body.Category == "" {
		body.Category = "Uncategorized"	
	}

	// Find or create category
	category, err := FindOrCreateCategory(body.Category, user.ID)
	if err != nil {
		http.Error(w, "Failed to process category", http.StatusInternalServerError)
		return
	}

	// Create transaction
	transaction := models.Transaction{
		Amount: body.Amount,
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

func ReadTransaction(w http.ResponseWriter, r *http.Request) {
	// Check for method
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get user info from middleware
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Type to store transaction
	var transactions []models.Transaction

	// Create time filter for later

	// Read the data from database
	result := initializers.DB.Limit(10).Where("user_id = ?", user.ID).Find(&transactions)

	if result.Error != nil {
		http.Error(w, "Failed to retrieve transactions!", http.StatusInternalServerError)
		return
	}

	// Return in json
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(transactions)
}

// Usage: https/link/updateTransaction
func UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	// Check for method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get user data from middleware
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		TransactionID int  `json:"transaction_id"`
		Amount        *int `json:"amount"`
		//Category      *string `json:"category"`
		Description *string `json:"description"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	// Create placeholder for existing transaction
	var existingTransaction models.Transaction

	result := initializers.DB.Where("id = ? AND user_id = ?", uint(body.TransactionID), user.ID).First(&existingTransaction)

	if result.Error != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	// Update value
	if body.Amount != nil {
		existingTransaction.Amount = *body.Amount
	}

	//if body.Category != nil {
	//	existingTransaction.Category = *body.Category
	//}

	if body.Description != nil {
		existingTransaction.Description = *body.Description
	}

	// Save updated data to database
	result = initializers.DB.Save(&existingTransaction)

	if result.Error != nil {
		http.Error(w, "Failed to update data", http.StatusInternalServerError)
		return
	}

	// Return in json
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(existingTransaction)
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
