package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"survey-battle-backend-go/internal/models"
	"survey-battle-backend-go/internal/service"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type RegisterRequest struct {
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	User        UserResponse `json:"user"`
	AccessToken string       `json:"accessToken"`
}

type UserResponse struct {
	ID          string             `json:"id"`
	Profile     models.Profile     `json:"profile"`
	EmailDetail models.EmailDetail `json:"emailDetail"`
	Role        string             `json:"role"`
}

type AuthHandler struct {
	AuthService *service.AuthService
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	input.FullName = strings.TrimSpace(input.FullName)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if err := validate.Struct(input); err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	user, accessToken, refreshToken, err :=
		h.AuthService.Register(r.Context(), input.FullName, input.Email, input.Password)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	setRefreshCookie(w, refreshToken)

	response := AuthResponse{
		User: UserResponse{
			ID:          user.ID.Hex(),
			Profile:     user.Profile,
			EmailDetail: user.EmailDetail,
			Role:        user.Role,
		},
		AccessToken: accessToken,
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if err := validate.Struct(input); err != nil {
		http.Error(w, "Invalid email or password", http.StatusBadRequest)
		return
	}

	user, accessToken, refreshToken, err :=
		h.AuthService.Login(r.Context(), input.Email, input.Password)

	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	setRefreshCookie(w, refreshToken)

	response := AuthResponse{
		User: UserResponse{
			ID:          user.ID.Hex(),
			Profile:     user.Profile,
			EmailDetail: user.EmailDetail,
			Role:        user.Role,
		},
		AccessToken: accessToken,
	}

	writeJSON(w, http.StatusOK, response)
}

func setRefreshCookie(w http.ResponseWriter, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
