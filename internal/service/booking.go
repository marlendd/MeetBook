package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-marlendd/internal/model"
)

type BookingService struct {
	bookingRepo BookingRepo
	slotRepo    SlotRepo
	log         *slog.Logger
}

func NewBookingService(bookingRepo BookingRepo, slotRepo SlotRepo, log *slog.Logger) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		slotRepo:    slotRepo,
		log:         log,
	}
}

func (b *BookingService) Create(ctx context.Context, userID, slotID uuid.UUID) (*model.Booking, error) {
	// получаем слот
	slot, err := b.slotRepo.GetByID(ctx, slotID)
	if err != nil {
		return nil, err
	}

	// проверяем наличие слота
	if slot == nil {
		b.log.Info("slot not found", "slotID", slotID)
		return nil, model.ErrSlotNotFound
	}

	// проверяем время начала слота
	if slot.Start.Before(time.Now()) {
		b.log.Info("start in past")
		return nil, model.ErrSlotInPast
	}

	booking := &model.Booking{
		SlotID: slotID,
		UserID: userID,
		Status: model.StatusActive,
	}

	if err := b.bookingRepo.Create(ctx, booking); err != nil {
		return nil, err
	}

	return booking, nil
}

func (b *BookingService) Cancel(ctx context.Context, userID, bookingID uuid.UUID) (*model.Booking, error) {
	booking, err := b.bookingRepo.GetById(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, model.ErrBookingNotFound
	}
	if booking.UserID != userID {
		return nil, model.ErrForbidden
	}
	if booking.Status == model.StatusCancelled {
		return booking, nil
	}

	if err := b.bookingRepo.Cancel(ctx, bookingID); err != nil {
		return nil, err
	}
	booking.Status = model.StatusCancelled

	return booking, nil
}

func (b *BookingService) ListAll(ctx context.Context, page, pageSize int) ([]model.Booking, int, error) {
	return b.bookingRepo.ListAll(ctx, page, pageSize)
}

func (b *BookingService) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Booking, error) {
	return b.bookingRepo.ListByUser(ctx, userID)
}
