package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"time"

	"github.com/DiedrickD/llm-powered-finance-tracker/initializers"
	"github.com/DiedrickD/llm-powered-finance-tracker/models"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get cookie
		cookie, err := r.Cookie("Authorization")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate JWT
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("SECRET_JWT")), nil
		})

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// check exp
			if float64(time.Now().Unix()) > claims["exp"].(float64) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Find the user with token sub
			var user models.User
			initializers.DB.First(&user, claims["sub"])

			if user.ID == 0 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user", user)

			// Success then call next handler
			next(w, r.WithContext(ctx))

		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}
}
