package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-marlendd/internal/model"
	"github.com/internships-backend/test-backend-marlendd/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleCreate_Success(t *testing.T) {
	roomID := uuid.New()
	room := &model.Room{ID: roomID, Name: "Room 1"}

	svc := service.NewScheduleService(
		&mockRoomRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Room, error) {
				return room, nil
			},
		},
		&mockScheduleRepo{
			getByRoomIdFn: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return nil, nil
			},
			createFn: func(_ context.Context, s *model.Schedule) error {
				s.ID = uuid.New()
				return nil
			},
		},
		testLog,
	)

	schedule := &model.Schedule{
		RoomID:     roomID,
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}

	err := svc.Create(context.Background(), roomID, schedule)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, schedule.ID)
}

func TestScheduleCreate_RoomNotFound(t *testing.T) {
	svc := service.NewScheduleService(
		&mockRoomRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Room, error) {
				return nil, nil
			},
		},
		&mockScheduleRepo{},
		testLog,
	)

	err := svc.Create(context.Background(), uuid.New(), &model.Schedule{})
	assert.ErrorIs(t, err, model.ErrRoomNotFound)
}

func TestScheduleCreate_AlreadyExists(t *testing.T) {
	roomID := uuid.New()

	svc := service.NewScheduleService(
		&mockRoomRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Room, error) {
				return &model.Room{ID: roomID}, nil
			},
		},
		&mockScheduleRepo{
			getByRoomIdFn: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{ID: uuid.New(), RoomID: roomID}, nil
			},
		},
		testLog,
	)

	err := svc.Create(context.Background(), roomID, &model.Schedule{RoomID: roomID})
	assert.ErrorIs(t, err, model.ErrScheduleExists)
}
