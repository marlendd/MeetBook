package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/internships-backend/test-backend-marlendd/internal/model"
)

type ConferenceClient interface {
	CreateLink(ctx context.Context, bookingID uuid.UUID) (string, error)
}

type BookingRepo interface {
	Create(ctx context.Context, booking *model.Booking) error
	GetById(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	Cancel(ctx context.Context, id uuid.UUID) error
	UpdateConferenceLink(ctx context.Context, id uuid.UUID, link string) error
	ListAll(ctx context.Context, page, pageSize int) ([]model.Booking, int, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Booking, error)
}

type SlotRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Slot, error)
	BulkUpsert(ctx context.Context, slots []model.Slot) error
	ListAvailable(ctx context.Context, roomID uuid.UUID, date time.Time) ([]model.Slot, error)
}

type RoomRepo interface {
	Create(ctx context.Context, room *model.Room) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Room, error)
	List(ctx context.Context) ([]model.Room, error)
}

type ScheduleRepo interface {
	Create(ctx context.Context, schedule *model.Schedule) error
	GetByRoomId(ctx context.Context, roomID uuid.UUID) (*model.Schedule, error)
}

type UserRepo interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}
