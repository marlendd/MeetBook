package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/internships-backend/test-backend-marlendd/internal/httputil"
	"github.com/internships-backend/test-backend-marlendd/internal/model"
	"github.com/internships-backend/test-backend-marlendd/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	log         *slog.Logger
}

func NewAuthHandler(authService *service.AuthService, log *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		log:         log,
	}
}

// DummyLogin godoc
//
//	@Summary		Получить тестовый JWT по роли
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dummyLoginRequest	true	"Роль"
//	@Success		200		{object}	tokenResponse
//	@Failure		400		{object}	httputil.ErrorResponse
//	@Failure		500		{object}	httputil.ErrorResponse
//	@Router			/dummyLogin [post]
func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.log.Error("failed to close request body", "error", err)
		}
	}()

	var req struct {
		Role model.Role `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	token, err := h.authService.DummyLogin(req.Role)
	if err != nil {
		if errors.Is(err, model.ErrInvalidRole) {
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid role")
		} else {
			h.log.Error("dummyLogin failed", "error", err)
			httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}

// Register godoc
//
//	@Summary		Регистрация пользователя
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		registerRequest	true	"Данные пользователя"
//	@Success		201		{object}	userResponse
//	@Failure		400		{object}	httputil.ErrorResponse
//	@Failure		500		{object}	httputil.ErrorResponse
//	@Router			/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.log.Error("failed to close request body", "error", err)
		}
	}()

	var req struct {
		Email    string     `json:"email"`
		Password string     `json:"password"`
		Role     model.Role `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "email and password are required")
		return
	}

	if req.Role != model.RoleAdmin && req.Role != model.RoleUser {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "role must be admin or user")
		return
	}

	user, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, model.ErrEmailTaken) {
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "email already taken")
		} else {
			h.log.Error("register failed", "error", err)
			httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]any{"user": user})
}

// Login godoc
//
//	@Summary		Авторизация по email и паролю
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		loginRequest	true	"Учётные данные"
//	@Success		200		{object}	tokenResponse
//	@Failure		401		{object}	httputil.ErrorResponse
//	@Failure		500		{object}	httputil.ErrorResponse
//	@Router			/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.log.Error("failed to close request body", "error", err)
		}
	}()

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, model.ErrInvalidCredentials) {
			httputil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		} else {
			h.log.Error("login failed", "error", err)
			httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}

// Request/response types for swagger docs

type dummyLoginRequest struct {
	Role string `json:"role" example:"user" enums:"admin,user"`
}

type registerRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secret123"`
	Role     string `json:"role" example:"user" enums:"admin,user"`
}

type loginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secret123"`
}

type tokenResponse struct {
	Token string `json:"token" example:"eyJhbGci..."`
}

type userResponse struct {
	User model.User `json:"user"`
}
