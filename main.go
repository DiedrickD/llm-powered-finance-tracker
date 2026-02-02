package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/DiedrickD/llm-powered-finance-tracker/controllers"
	initializers "github.com/DiedrickD/llm-powered-finance-tracker/initializers"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sqlDB, err := initializers.DB.DB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := sqlDB.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func main() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
	initializers.SyncDatabase()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/signup", controllers.Signup)

	server := &http.Server{
		Addr:           ":3000",
		Handler:        mux,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("Server running on :3000")
	log.Fatal(server.ListenAndServe())
}
