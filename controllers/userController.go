package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DiedrickD/llm-powered-finance-tracker/initializers"
	"github.com/DiedrickD/llm-powered-finance-tracker/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Signup(w http.ResponseWriter, r *http.Request) {
	// Get the post request
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	if body.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if body.Email == "" || body.Password == "" {
		http.Error(w, "Email and password required", http.StatusBadRequest)
		return
	}

	// Check for existing username
	var existingUser models.User

	err = initializers.DB.Where("username = ?", body.Username).First(&existingUser).Error
	if err == nil {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Check for existing email
	var existingEmail models.User

	err = initializers.DB.Where("email = ?", body.Email).First(&existingEmail).Error
	if err == nil {
		http.Error(w, "Email already used!", http.StatusConflict)
		return
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Hash the password
	cost := bcrypt.DefaultCost

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), cost)
	if err != nil {
		http.Error(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	// Create User
	newUser := models.User{
		Username: body.Username,
		Email:    body.Email,
		Password: string(hashedPassword),
		Currency: "IDR", // Todo: Give option to change later
		NewUser:  true,
	}

	result := initializers.DB.Create(&newUser)
	if result.Error != nil {
		http.Error(w, "Username or email already exists", http.StatusConflict)
		return
	}

	// Respond
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "user created",
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	// Get the email and password
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	// Look up in databasev
	var user models.User

	result := initializers.DB.First(&user, "email = ?", body.Email)

	// Check if the user is not found
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Invalid email or password", http.StatusBadRequest)
			return
		}

		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Compare password hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))

	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusBadRequest)
		return
	}

	// Generate jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret key
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_JWT")))

	if err != nil {
		http.Error(w, "Failed to create token", http.StatusBadRequest)
		return
	}

	// Create cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	// Send it back
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "logged in successfully",
		"user": map[string]string{
			"email": user.Email,
		},
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// Set maxage cookie to -1 to logout
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})

}

func Validate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user := r.Context().Value("user").(models.User)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "I'm logged in",
		"user":    user.Email,
	})
}

func UpdateCurrency(w http.ResponseWriter, r *http.Request) {
	// Method check
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Currency string `json:"currency"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	if len(body.Currency) != 3 {
		http.Error(w, "Currency must be a 3-letter code (e.g., IDR, USD)", http.StatusBadRequest)
		return
	}
	// Todo : check for the money type or list it maybe??

	newCurrency := strings.ToUpper(body.Currency)

	result := initializers.DB.Model(&user).Update("currency", newCurrency)

	if result.Error != nil {
		http.Error(w, "Failed to update currency", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Currency updated successfully",
		"currency": newCurrency,
	})
}

func FirstBalanceInput(w http.ResponseWriter, r *http.Request) {
	// Method check
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (auth middleware)
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only allowed once
	if !user.NewUser {
		http.Error(w, "Initial balance already set", http.StatusBadRequest)
		return
	}

	var body struct {
		Balance int `json:"balance"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Balance < 0 {
		http.Error(w, "Balance cannot be negative", http.StatusBadRequest)
		return
	}

	// Update balance & mark user as not new
	result := initializers.DB.Model(&user).Updates(map[string]interface{}{
		"balance":  body.Balance,
	})

	if result.Error != nil {
		http.Error(w, "Failed to set initial balance", http.StatusInternalServerError)
		return
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Initial balance set successfully",
		"balance": body.Balance,
	})
}

func ChangeBalance(w http.ResponseWriter, r *http.Request) {
	// Method check
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get user from context
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Balance int `json:"balance"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Balance < 0 {
		http.Error(w, "Balance cannot be negative", http.StatusBadRequest)
		return
	}

	// Update balance
	result := initializers.DB.Model(&user).Update("balance", body.Balance)
	if result.Error != nil {
		http.Error(w, "Failed to update balance", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Balance updated successfully",
		"balance": body.Balance,
	})
}
