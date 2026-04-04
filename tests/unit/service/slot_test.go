package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/internships-backend/test-backend-marlendd/internal/model"
	"github.com/internships-backend/test-backend-marlendd/internal/service"
)

func TestSlotListAvailable_Success(t *testing.T) {
	roomID := uuid.New()
	date := time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC) // понедельник

	expectedSlots := []model.Slot{
		{ID: uuid.New(), RoomID: roomID, Start: date.Add(9 * time.Hour), End: date.Add(9*time.Hour + 30*time.Minute)},
	}

	svc := service.NewSlotService(
		&mockSlotRepo{
			bulkUpsertFn:    func(_ context.Context, _ []model.Slot) error { return nil },
			listAvailableFn: func(_ context.Context, _ uuid.UUID, _ time.Time) ([]model.Slot, error) { return expectedSlots, nil },
		},
		&mockScheduleRepo{
			getByRoomIdFn: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{
					RoomID:     roomID,
					DaysOfWeek: []int{1, 2, 3, 4, 5, 6, 7},
					StartTime:  "09:00",
					EndTime:    "18:00",
				}, nil
			},
		},
		&mockRoomRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Room, error) {
				return &model.Room{ID: roomID}, nil
			},
		},
		testLog,
	)

	slots, err := svc.ListAvailable(context.Background(), roomID, date)
	require.NoError(t, err)
	assert.Len(t, slots, 1)
}

func TestSlotListAvailable_RoomNotFound(t *testing.T) {
	svc := service.NewSlotService(
		&mockSlotRepo{},
		&mockScheduleRepo{},
		&mockRoomRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Room, error) { return nil, nil },
		},
		testLog,
	)

	_, err := svc.ListAvailable(context.Background(), uuid.New(), time.Now())
	assert.ErrorIs(t, err, model.ErrRoomNotFound)
}

func TestSlotListAvailable_NoSchedule(t *testing.T) {
	roomID := uuid.New()

	svc := service.NewSlotService(
		&mockSlotRepo{},
		&mockScheduleRepo{
			getByRoomIdFn: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) { return nil, nil },
		},
		&mockRoomRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Room, error) {
				return &model.Room{ID: roomID}, nil
			},
		},
		testLog,
	)

	slots, err := svc.ListAvailable(context.Background(), roomID, time.Now())
	require.NoError(t, err)
	assert.Empty(t, slots)
}

func TestSlotListAvailable_WrongDay(t *testing.T) {
	roomID := uuid.New()
	sunday := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC) // воскресенье

	svc := service.NewSlotService(
		&mockSlotRepo{},
		&mockScheduleRepo{
			getByRoomIdFn: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{
					RoomID:     roomID,
					DaysOfWeek: []int{1, 2, 3, 4, 5}, // только пн-пт
					StartTime:  "09:00",
					EndTime:    "18:00",
				}, nil
			},
		},
		&mockRoomRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Room, error) {
				return &model.Room{ID: roomID}, nil
			},
		},
		testLog,
	)

	slots, err := svc.ListAvailable(context.Background(), roomID, sunday)
	require.NoError(t, err)
	assert.Empty(t, slots)
}

func TestSlotListAvailable_GeneratesCorrectCount(t *testing.T) {
	roomID := uuid.New()
	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC) // понедельник

	var upserted []model.Slot

	svc := service.NewSlotService(
		&mockSlotRepo{
			bulkUpsertFn: func(_ context.Context, slots []model.Slot) error {
				upserted = slots
				return nil
			},
			listAvailableFn: func(_ context.Context, _ uuid.UUID, _ time.Time) ([]model.Slot, error) {
				return upserted, nil
			},
		},
		&mockScheduleRepo{
			getByRoomIdFn: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{
					RoomID:     roomID,
					DaysOfWeek: []int{1, 2, 3, 4, 5, 6, 7},
					StartTime:  "09:00",
					EndTime:    "18:00",
				}, nil
			},
		},
		&mockRoomRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Room, error) {
				return &model.Room{ID: roomID}, nil
			},
		},
		testLog,
	)

	slots, err := svc.ListAvailable(context.Background(), roomID, date)
	require.NoError(t, err)
	// 9:00-18:00 = 9 часов = 18 слотов по 30 минут
	assert.Len(t, slots, 18)
}
