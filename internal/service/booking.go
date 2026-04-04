package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/internships-backend/test-backend-marlendd/internal/model"
)

type BookingService struct {
	bookingRepo      BookingRepo
	slotRepo         SlotRepo
	conferenceClient ConferenceClient
	log              *slog.Logger
}

func NewBookingService(bookingRepo BookingRepo, slotRepo SlotRepo, conferenceClient ConferenceClient, log *slog.Logger) *BookingService {
	return &BookingService{
		bookingRepo:      bookingRepo,
		slotRepo:         slotRepo,
		conferenceClient: conferenceClient,
		log:              log,
	}
}

func (b *BookingService) Create(ctx context.Context, userID, slotID uuid.UUID, createConferenceLink bool) (*model.Booking, error) {
	slot, err := b.slotRepo.GetByID(ctx, slotID)
	if err != nil {
		return nil, err
	}
	if slot == nil {
		b.log.Info("slot not found", "slotID", slotID)
		return nil, model.ErrSlotNotFound
	}
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

	// Запрашиваем ссылку на конференцию после успешного создания брони.
	// Сбой внешнего сервиса не откатывает бронь — она уже создана.
	// Ссылка просто не будет заполнена, что допустимо по бизнес-логике.
	if createConferenceLink && b.conferenceClient != nil {
		link, err := b.conferenceClient.CreateLink(ctx, booking.ID)
		if err != nil {
			b.log.Warn("failed to create conference link, booking created without it",
				"bookingID", booking.ID, "error", err)
		} else {
			booking.ConferenceLink = &link
			if err := b.bookingRepo.UpdateConferenceLink(ctx, booking.ID, link); err != nil {
				b.log.Warn("failed to save conference link",
					"bookingID", booking.ID, "error", err)
				booking.ConferenceLink = nil
			}
		}
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
	if booking.Status == model.StatusCanceled {
		return booking, nil
	}

	if err := b.bookingRepo.Cancel(ctx, bookingID); err != nil {
		return nil, err
	}
	booking.Status = model.StatusCanceled

	return booking, nil
}

func (b *BookingService) ListAll(ctx context.Context, page, pageSize int) ([]model.Booking, int, error) {
	return b.bookingRepo.ListAll(ctx, page, pageSize)
}

func (b *BookingService) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Booking, error) {
	return b.bookingRepo.ListByUser(ctx, userID)
}
