package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-marlendd/internal/httputil"
	"github.com/internships-backend/test-backend-marlendd/internal/model"
	"github.com/internships-backend/test-backend-marlendd/internal/service"
)

type ScheduleHandler struct {
	scheduleService *service.ScheduleService
	log             *slog.Logger
}

func NewScheduleHandler(scheduleService *service.ScheduleService, log *slog.Logger) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleService: scheduleService,
		log:             log,
	}
}

func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var schedule struct {
		DaysOfWeek []int  `json:"daysOfWeek"`
		StartTime  string `json:"startTime"`
		EndTime    string `json:"endTime"`
	}

	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		h.log.Error("failed to decode body", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	for _, d := range schedule.DaysOfWeek {
		if d < 1 || d > 7 {
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "daysOfWeek must be between 1 and 7")
			return
		}
	}

	newSchedule := &model.Schedule{
		RoomID:     roomID,
		DaysOfWeek: schedule.DaysOfWeek,
		StartTime:  schedule.StartTime,
		EndTime:    schedule.EndTime,
	}

	if err = h.scheduleService.Create(r.Context(), roomID, newSchedule); err != nil {
		switch {
		case errors.Is(err, model.ErrRoomNotFound):
			httputil.WriteError(w, http.StatusNotFound, "ROOM_NOT_FOUND", "room not found")
		case errors.Is(err, model.ErrScheduleExists):
			httputil.WriteError(w, http.StatusConflict, "SCHEDULE_EXISTS", "schedule for this room already exists and cannot be changed")
		default:
			h.log.Error("failed to create schedule", "error", err)
			httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]any{"schedule": newSchedule})
}
