package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-marlendd/internal/model"
)

type BookingRepo interface {
	Create(ctx context.Context, booking *model.Booking) error
	GetById(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	Cancel(ctx context.Context, id uuid.UUID) error
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
