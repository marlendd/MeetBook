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
