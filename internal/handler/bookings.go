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

// Create godoc
//
//	@Summary		Создать бронь на слот
//	@Tags			Bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createBookingRequest	true	"Данные брони"
//	@Success		201		{object}	bookingResponse
//	@Failure		400		{object}	httputil.ErrorResponse
//	@Failure		401		{object}	httputil.ErrorResponse
//	@Failure		403		{object}	httputil.ErrorResponse
//	@Failure		404		{object}	httputil.ErrorResponse
//	@Failure		409		{object}	httputil.ErrorResponse
//	@Failure		500		{object}	httputil.ErrorResponse
//	@Router			/bookings/create [post]
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

// Cancel godoc
//
//	@Summary		Отменить бронь
//	@Tags			Bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			bookingId	path		string	true	"ID брони"	format(uuid)
//	@Success		200			{object}	bookingResponse
//	@Failure		401			{object}	httputil.ErrorResponse
//	@Failure		403			{object}	httputil.ErrorResponse
//	@Failure		404			{object}	httputil.ErrorResponse
//	@Failure		500			{object}	httputil.ErrorResponse
//	@Router			/bookings/{bookingId}/cancel [post]
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

// ListByUser godoc
//
//	@Summary		Мои брони
//	@Tags			Bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	bookingsResponse
//	@Failure		401	{object}	httputil.ErrorResponse
//	@Failure		403	{object}	httputil.ErrorResponse
//	@Failure		500	{object}	httputil.ErrorResponse
//	@Router			/bookings/my [get]
func (h *BookingHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	bookings, err := h.bookingService.ListByUser(r.Context(), userID)
	if err != nil {
		h.log.Error("failed to list bookings", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	if bookings == nil {
		bookings = []model.Booking{}
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{"bookings": bookings})
}

// ListAll godoc
//
//	@Summary		Список всех броней с пагинацией
//	@Tags			Bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Номер страницы"	default(1)
//	@Param			pageSize	query		int	false	"Размер страницы"	default(20)
//	@Success		200			{object}	bookingsPageResponse
//	@Failure		400			{object}	httputil.ErrorResponse
//	@Failure		401			{object}	httputil.ErrorResponse
//	@Failure		403			{object}	httputil.ErrorResponse
//	@Failure		500			{object}	httputil.ErrorResponse
//	@Router			/bookings/list [get]
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
		h.log.Error("failed to list all bookings", "error", err)
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

type createBookingRequest struct {
	SlotID               string `json:"slotId" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreateConferenceLink bool   `json:"createConferenceLink" example:"false"`
}

type bookingResponse struct {
	Booking model.Booking `json:"booking"`
}

type bookingsResponse struct {
	Bookings []model.Booking `json:"bookings"`
}

type paginationInfo struct {
	Page     int `json:"page" example:"1"`
	PageSize int `json:"pageSize" example:"20"`
	Total    int `json:"total" example:"100"`
}

type bookingsPageResponse struct {
	Bookings   []model.Booking `json:"bookings"`
	Pagination paginationInfo  `json:"pagination"`
}
