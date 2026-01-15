package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var db *sql.DB = createDB()

func main() {
	r := chi.NewRouter()
	// r.Use(middleware.Logger)
	// r.Use(middleware.Recoverer)

	r.Post("/signup", signUpUser)
	r.Post("/login", loginUser)
	// r.Get("/users", listUsers)
	// r.Post("/users", createUser)
	// r.Get("/users/{id}", getUser)
	// r.Put("/users/{id}", updateUser)
	// r.Delete("/users/{id}", deleteUser)

	fmt.Println("Listening on port 8080")
	http.ListenAndServe(":8080", r)
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	username := req.Username
	hash := md5.Sum([]byte(req.Password))
	login := checkIfValidLogin(db, username, hash)
	if login {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Login successful"))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Login failed"))
	}

}

func signUpUser(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	username := req.Username
	hash := md5.Sum([]byte(req.Password))
	operation := addUser(db, username, hash)
	if operation {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Signup successful"))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Signup failed"))
	}
}
func getUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user := User{ID: id, Name: "John", Email: "john@example.com"}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Implement other handlers...
