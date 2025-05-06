package handler

import (
	"errors"
	"net/http"
	"time"

	"real-time-forum/backend/internal/config"
	domain "real-time-forum/backend/internal/model"
	"real-time-forum/backend/internal/repository"
	"real-time-forum/backend/internal/utils"
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
	var req domain.RegisterRequest
	if err := utils.DecodeJSONBody(w, r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := utils.Validate.Struct(req); err != nil {
		utils.RespondWithValidationError(w, err)
		return
	}

	exists, err := h.userRepo.EmailOrNicknameExists(req.Email, req.Nickname)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not check user existence")
		return
	}
	if exists {
		utils.RespondWithError(w, http.StatusConflict, "User already exists")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password, h.cfg.BcryptCost)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}

	user := domain.User{
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	token, expiresAt, err := h.sessionRepo.Create(int64(createdUser.ID), h.cfg.SessionTimeout)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not create session")
		return
	}

	utils.SetAuthCookie(w, token, expiresAt, h.cfg.IsProduction())

	utils.RespondWithJSON(w, http.StatusCreated, domain.AuthResponse{
		User:      createdUser.ToDTO(),
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := utils.DecodeJSONBody(w, r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	user, err := h.userRepo.FindByIdentifier(req.Identifier)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			_ = utils.ComparePasswords("$2a$10$fakehash", req.Password) // Prevent timing attack
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not find user")
		return
	}

	if !utils.ComparePasswords(user.PasswordHash, req.Password) {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, expiresAt, err := h.sessionRepo.Create(int64(user.ID), h.cfg.SessionTimeout)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not create session")
		return
	}

	utils.SetAuthCookie(w, token, expiresAt, h.cfg.IsProduction())

	utils.RespondWithJSON(w, http.StatusOK, domain.AuthResponse{
		User:      user.ToDTO(),
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := utils.GetAuthToken(r)
	if token == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing token")
		return
	}

	if err := h.sessionRepo.Delete(token); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not logout")
		return
	}

	utils.ClearAuthCookie(w, h.cfg.IsProduction())
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Logged out"})
}
