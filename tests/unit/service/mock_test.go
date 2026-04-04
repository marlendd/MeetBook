package service_test

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/internships-backend/test-backend-marlendd/internal/model"
)

// --- BookingRepo mock ---

type mockBookingRepo struct {
	createFn     func(ctx context.Context, booking *model.Booking) error
	getByIdFn    func(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	cancelFn     func(ctx context.Context, id uuid.UUID) error
	listAllFn    func(ctx context.Context, page, pageSize int) ([]model.Booking, int, error)
	listByUserFn func(ctx context.Context, userID uuid.UUID) ([]model.Booking, error)
}

func (m *mockBookingRepo) Create(ctx context.Context, booking *model.Booking) error {
	return m.createFn(ctx, booking)
}
func (m *mockBookingRepo) GetById(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	return m.getByIdFn(ctx, id)
}
func (m *mockBookingRepo) Cancel(ctx context.Context, id uuid.UUID) error {
	return m.cancelFn(ctx, id)
}
func (m *mockBookingRepo) ListAll(ctx context.Context, page, pageSize int) ([]model.Booking, int, error) {
	return m.listAllFn(ctx, page, pageSize)
}
func (m *mockBookingRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Booking, error) {
	return m.listByUserFn(ctx, userID)
}

// --- SlotRepo mock ---

type mockSlotRepo struct {
	getByIDFn       func(ctx context.Context, id uuid.UUID) (*model.Slot, error)
	bulkUpsertFn    func(ctx context.Context, slots []model.Slot) error
	listAvailableFn func(ctx context.Context, roomID uuid.UUID, date time.Time) ([]model.Slot, error)
}

func (m *mockSlotRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Slot, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockSlotRepo) BulkUpsert(ctx context.Context, slots []model.Slot) error {
	return m.bulkUpsertFn(ctx, slots)
}
func (m *mockSlotRepo) ListAvailable(ctx context.Context, roomID uuid.UUID, date time.Time) ([]model.Slot, error) {
	return m.listAvailableFn(ctx, roomID, date)
}

// --- RoomRepo mock ---

type mockRoomRepo struct {
	createFn  func(ctx context.Context, room *model.Room) error
	getByIDFn func(ctx context.Context, id uuid.UUID) (*model.Room, error)
	listFn    func(ctx context.Context) ([]model.Room, error)
}

func (m *mockRoomRepo) Create(ctx context.Context, room *model.Room) error {
	return m.createFn(ctx, room)
}
func (m *mockRoomRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Room, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockRoomRepo) List(ctx context.Context) ([]model.Room, error) {
	return m.listFn(ctx)
}

// --- ScheduleRepo mock ---

type mockScheduleRepo struct {
	createFn      func(ctx context.Context, schedule *model.Schedule) error
	getByRoomIdFn func(ctx context.Context, roomID uuid.UUID) (*model.Schedule, error)
}

func (m *mockScheduleRepo) Create(ctx context.Context, schedule *model.Schedule) error {
	return m.createFn(ctx, schedule)
}
func (m *mockScheduleRepo) GetByRoomId(ctx context.Context, roomID uuid.UUID) (*model.Schedule, error) {
	return m.getByRoomIdFn(ctx, roomID)
}
