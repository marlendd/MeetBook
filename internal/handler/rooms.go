package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/internships-backend/test-backend-marlendd/internal/httputil"
	"github.com/internships-backend/test-backend-marlendd/internal/model"
	"github.com/internships-backend/test-backend-marlendd/internal/service"
)

type RoomHandler struct {
	roomService *service.RoomService
	log         *slog.Logger
}

func NewRoomHandler(roomService *service.RoomService, log *slog.Logger) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
		log:         log,
	}
}

func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.log.Error("failed to close request body", "error", err)
		}
	}()

	var room struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Capacity    *int    `json:"capacity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&room); err != nil {
		h.log.Error("failed to decode body", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	newRoom := &model.Room{
		Name:        room.Name,
		Description: room.Description,
		Capacity:    room.Capacity,
	}

	if err := h.roomService.Create(r.Context(), newRoom); err != nil {
		h.log.Error("failed to create room", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]any{"room": newRoom})
}

func (h *RoomHandler) List(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.roomService.List(r.Context())
	if err != nil {
		h.log.Error("failed to list rooms", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	if rooms == nil {
		rooms = []model.Room{}
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"rooms": rooms})
}
