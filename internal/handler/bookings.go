package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/internships-backend/test-backend-marlendd/internal/httputil"
	"github.com/internships-backend/test-backend-marlendd/internal/middleware"
	"github.com/internships-backend/test-backend-marlendd/internal/model"
	"github.com/internships-backend/test-backend-marlendd/internal/service"
)

type BookingHandler struct {
	bookingService *service.BookingService
	log            *slog.Logger
}

func NewBookingHandler(bookingService *service.BookingService, log *slog.Logger) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
		log:            log,
	}
}

func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.log.Error("failed to close request body", "error", err)
		}
	}()

	var req struct {
		SlotID               uuid.UUID `json:"slotId"`
		CreateConferenceLink bool      `json:"createConferenceLink"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode body", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	userID := middleware.GetUserID(r.Context())

	booking, err := h.bookingService.Create(r.Context(), userID, req.SlotID, req.CreateConferenceLink)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrSlotNotFound):
			httputil.WriteError(w, http.StatusNotFound, "SLOT_NOT_FOUND", "slot not found")
		case errors.Is(err, model.ErrSlotInPast):
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "slot is in the past")
		case errors.Is(err, model.ErrSlotAlreadyBooked):
			httputil.WriteError(w, http.StatusConflict, "SLOT_ALREADY_BOOKED", "slot is already booked")
		default:
			h.log.Error("failed to create booking", "error", err)
			httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]any{"booking": booking})
}

func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := r.PathValue("bookingId")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		h.log.Error("failed to parse bookingId", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())

	booking, err := h.bookingService.Cancel(r.Context(), userID, bookingID)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrBookingNotFound):
			httputil.WriteError(w, http.StatusNotFound, "BOOKING_NOT_FOUND", "booking not found")
		case errors.Is(err, model.ErrForbidden):
			httputil.WriteError(w, http.StatusForbidden, "FORBIDDEN", "cannot cancel another user's booking")
		default:
			h.log.Error("failed to cancel booking", "error", err)
			httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{"booking": booking})
}

func (h *BookingHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	bookings, err := h.bookingService.ListByUser(r.Context(), userID)
	if err != nil {
		h.log.Error("failed to cancel booking", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	if bookings == nil {
		bookings = []model.Booking{}
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{"bookings": bookings})
}

func (h *BookingHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	var err error
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if page, err = strconv.Atoi(p); err != nil || page < 1 {
			h.log.Error("failed to parse page", "error", err)
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid page")
			return
		}
	}

	pageSize := 20
	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if pageSize, err = strconv.Atoi(ps); err != nil || pageSize < 1 || pageSize > 100 {
			h.log.Error("failed to parse page size", "error", err)
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid page size")
			return
		}
	}

	bookings, total, err := h.bookingService.ListAll(r.Context(), page, pageSize)
	if err != nil {
		h.log.Error("failed to cancel booking", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	if bookings == nil {
		bookings = []model.Booking{}
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"bookings": bookings,
		"pagination": map[string]any{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
		},
	})
}
