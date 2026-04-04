package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/internships-backend/test-backend-marlendd/internal/httputil"
	"github.com/internships-backend/test-backend-marlendd/internal/model"
	"github.com/internships-backend/test-backend-marlendd/internal/service"
)

type SlotsHandler struct {
	slotsService *service.SlotService
	log          *slog.Logger
}

func NewSlotsHandler(SlotService *service.SlotService, log *slog.Logger) *SlotsHandler {
	return &SlotsHandler{
		slotsService: SlotService,
		log:          log,
	}
}

func (h *SlotsHandler) ListAvailable(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.log.Error("failed to close request body", "error", err)
		}
	}()

	roomIDStr := r.PathValue("roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		h.log.Error("failed to decode body", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "date is required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.log.Error("failed to decode date", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	slots, err := h.slotsService.ListAvailable(r.Context(), roomID, date)
	if err != nil {
		if errors.Is(err, model.ErrRoomNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "ROOM_NOT_FOUND", "room not found")
		} else {
			h.log.Error("failed to list slots", "error", err)
			httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	if slots == nil {
		slots = []model.Slot{}
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{"slots": slots})
}
