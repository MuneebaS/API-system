package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	JWTSecret    []byte
	DatabasePath string
}

type Server struct {
	db      *sql.DB
	config  Config
	queries *Queries
}

type RegisterRequest struct {
	Username         string `json:"username"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	SecurityQuestion string `json:"securityQuestion"`
	SecurityAnswer   string `json:"securityAnswer"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgotPasswordRequest struct {
	Email          string `json:"email"`
	SecurityAnswer string `json:"securityAnswer"`
	NewPassword    string `json:"newPassword"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func initDatabase(db *sql.DB) error {
	schema := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT UNIQUE NOT NULL,
        email TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        security_question TEXT NOT NULL,
        security_answer_hash TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    `
	_, err := db.Exec(schema)
	return err
}

func NewServer(config Config) (*Server, error) {
	db, err := sql.Open("sqlite3", config.DatabasePath)
	if err != nil {
		return nil, err
	}

	// Initialize the database
	if err = initDatabase(db); err != nil {
		return nil, err
	}

	return &Server{
		db:      db,
		config:  config,
		queries: New(db),
	}, nil
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Hash password and security answer
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	securityAnswerHash, err := bcrypt.GenerateFromPassword([]byte(req.SecurityAnswer), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user
	_, err = s.queries.CreateUser(context.Background(), CreateUserParams{
		Username:           req.Username,
		Email:              req.Email,
		PasswordHash:       string(passwordHash),
		SecurityQuestion:   req.SecurityQuestion,
		SecurityAnswerHash: string(securityAnswerHash),
	})

	if err != nil {
		log.Println(err)
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user by email
	user, err := s.queries.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.config.JWTSecret)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(LoginResponse{Token: tokenString})
}

func (s *Server) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user by email
	user, err := s.queries.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Verify security answer
	if err := bcrypt.CompareHashAndPassword([]byte(user.SecurityAnswerHash), []byte(req.SecurityAnswer)); err != nil {
		http.Error(w, "Invalid security answer", http.StatusUnauthorized)
		return
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update password
	err = s.queries.UpdateUserPassword(context.Background(), UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: string(newPasswordHash),
	})

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) authenticateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			return s.config.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.queries.ListUsers(context.Background())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(users)
}

func main() {
	config := Config{
		JWTSecret:    []byte(os.Getenv("JWT_SECRET")),
		DatabasePath: "users.db",
	}

	server, err := NewServer(config)
	if err != nil {
		log.Fatal(err)
	}
	defer server.db.Close()

	http.HandleFunc("POST /register", server.handleRegister)
	http.HandleFunc("POST /login", server.handleLogin)
	http.HandleFunc("POST /forgot-password", server.handleForgotPassword)
	http.HandleFunc("GET /users", server.authenticateMiddleware(server.handleListUsers))

	log.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
