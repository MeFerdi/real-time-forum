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
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Sanitize inputs
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Nickname = strings.TrimSpace(req.Nickname)

	// Validate inputs
	if err := utils.ValidateEmail(req.Email); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := utils.ValidatePassword(req.Password); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if email or nickname exists
	if user, err := h.userRepo.GetByEmail(req.Email); err != nil {
		writeError(w, "Internal server error", http.StatusInternalServerError)
		return
	} else if user != nil {
		writeError(w, "Email already exists", http.StatusConflict)
		return
	}

	if user, err := h.userRepo.GetByNickname(req.Nickname); err != nil {
		writeError(w, "Internal server error", http.StatusInternalServerError)
		return
	} else if user != nil {
		writeError(w, "Nickname already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		writeError(w, "Internal server error", http.StatusInternalServerError)
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

	if err := h.userRepo.Create(&user); err != nil {
		writeError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Fetch the created user to get the ID
	createdUser, err := h.userRepo.GetByEmail(req.Email)
	if err != nil || createdUser == nil {
		writeError(w, "Failed to fetch created user", http.StatusInternalServerError)
		return
	}

	session := model.Session{
		UserID:    createdUser.ID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(h.cfg.SessionTimeout),
		CreatedAt: time.Now(),
	}

	if err := h.sessionRepo.Create(&session); err != nil {
		writeError(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	utils.SetAuthCookie(w, session.Token, session.ExpiresAt, h.cfg.IsProduction())
	writeJSONResponse(w, http.StatusCreated, model.AuthResponse{
		User:      createdUser.ToDTO(),
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Identifier = strings.TrimSpace(strings.ToLower(req.Identifier))

	var user *model.User
	var err error
	if strings.Contains(req.Identifier, "@") {
		user, err = h.userRepo.GetByEmail(req.Identifier)
	} else {
		user, err = h.userRepo.GetByNickname(req.Identifier)
	}
	if err != nil || user == nil {
		writeError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !utils.ComparePasswords(user.PasswordHash, req.Password) {
		writeError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	session := model.Session{
		UserID:    user.ID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(h.cfg.SessionTimeout),
		CreatedAt: time.Now(),
	}

	if err := h.sessionRepo.Create(&session); err != nil {
		writeError(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	utils.SetAuthCookie(w, session.Token, session.ExpiresAt, h.cfg.IsProduction())
	writeJSONResponse(w, http.StatusOK, model.AuthResponse{
		User:      user.ToDTO(),
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := utils.GetAuthToken(r)
	if token == "" {
		writeError(w, "No authentication token", http.StatusBadRequest)
		return
	}

	if err := h.sessionRepo.Delete(token); err != nil {
		writeError(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	utils.ClearAuthCookie(w, h.cfg.IsProduction())
	w.WriteHeader(http.StatusOK)
}
