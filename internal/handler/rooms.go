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

// Create godoc
//
//	@Summary		Создать переговорку
//	@Tags			Rooms
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createRoomRequest	true	"Данные переговорки"
//	@Success		201		{object}	roomResponse
//	@Failure		400		{object}	httputil.ErrorResponse
//	@Failure		401		{object}	httputil.ErrorResponse
//	@Failure		403		{object}	httputil.ErrorResponse
//	@Failure		500		{object}	httputil.ErrorResponse
//	@Router			/rooms/create [post]
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

	if newRoom.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "name is required")
		return
	}

	if err := h.roomService.Create(r.Context(), newRoom); err != nil {
		h.log.Error("failed to create room", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]any{"room": newRoom})
}

// List godoc
//
//	@Summary		Список переговорок
//	@Tags			Rooms
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	roomsResponse
//	@Failure		401	{object}	httputil.ErrorResponse
//	@Failure		500	{object}	httputil.ErrorResponse
//	@Router			/rooms/list [get]
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

type createRoomRequest struct {
	Name        string  `json:"name" example:"Переговорка 1"`
	Description *string `json:"description" example:"Большая переговорка"`
	Capacity    *int    `json:"capacity" example:"8"`
}

type roomResponse struct {
	Room model.Room `json:"room"`
}

type roomsResponse struct {
	Rooms []model.Room `json:"rooms"`
}
