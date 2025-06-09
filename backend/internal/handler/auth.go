package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"real-time/backend/internal/config"
	"real-time/backend/internal/model"
	"real-time/backend/internal/repository"
	"real-time/backend/internal/utils"
)

type AuthHandler struct {
	cfg         *config.Config
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewAuthHandler(cfg *config.Config, userRepo repository.UserRepository, sessionRepo repository.SessionRepository) *AuthHandler {
	return &AuthHandler{
		cfg:         cfg,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Sanitize inputs
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Nickname = strings.TrimSpace(req.Nickname)

	// Validate inputs
	if err := utils.ValidateEmail(req.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := utils.ValidatePassword(req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	exists, err := h.userRepo.EmailOrNicknameExists(req.Email, req.Nickname)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := model.User{
		UUID:         uuid.New().String(),
		Nickname:     req.Nickname,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Age:          req.Age,
		Gender:       req.Gender,
		IsOnline:     true,
		LastOnline:   time.Now(),
	}

	createdUser, err := h.userRepo.Create(user)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	token, expiresAt, err := h.sessionRepo.Create(int64(createdUser.ID), h.cfg.SessionTimeout)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	utils.SetAuthCookie(w, token, expiresAt, h.cfg.IsProduction())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.AuthResponse{
		User:      createdUser.ToDTO(),
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Identifier = strings.TrimSpace(strings.ToLower(req.Identifier))

	user, err := h.userRepo.FindByIdentifier(req.Identifier)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !utils.ComparePasswords(user.PasswordHash, req.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := h.userRepo.SetOnlineStatus(user.ID, true); err != nil {
		http.Error(w, "Failed to update online status", http.StatusInternalServerError)
		return
	}

	token, expiresAt, err := h.sessionRepo.Create(int64(user.ID), h.cfg.SessionTimeout)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	utils.SetAuthCookie(w, token, expiresAt, h.cfg.IsProduction())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.AuthResponse{
		User:      user.ToDTO(),
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := utils.GetAuthToken(r)
	if token == "" {
		http.Error(w, "No authentication token", http.StatusBadRequest)
		return
	}

	userID, err := h.sessionRepo.Get(token)
	if err == nil && userID > 0 {
		if err := h.userRepo.SetOnlineStatus(int(userID), false); err != nil {
			http.Error(w, "Failed to update online status", http.StatusInternalServerError)
			return
		}
	}

	if err := h.sessionRepo.Delete(token); err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	utils.ClearAuthCookie(w, h.cfg.IsProduction())
	w.WriteHeader(http.StatusOK)
}