	package controllers

	import (
		"encoding/json"		
		"net/http"

		"github.com/DiedrickD/llm-powered-finance-tracker/initializers"
		"github.com/DiedrickD/llm-powered-finance-tracker/models"
		"golang.org/x/crypto/bcrypt"
	)

	func Signup(w http.ResponseWriter, r *http.Request) {
		// Get the post request 
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

		if body.Email == "" || body.Password == "" {
			http.Error(w, "Email and password required", http.StatusBadRequest)
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
			Email: body.Email,
			Password: string(hashedPassword),
		}

		initializers.DB.Create(&newUser)

		// Respond
		w.Header().Set("Content-Type", "application/json")			
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "user created",
		})

	}