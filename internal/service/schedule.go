package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-marlendd/internal/model"
)

type ScheduleService struct {
	roomRepo     RoomRepo
	scheduleRepo ScheduleRepo
	log          *slog.Logger
}

func NewScheduleService(roomRepo RoomRepo, scheduleRepo ScheduleRepo, log *slog.Logger) *ScheduleService {
	return &ScheduleService{
		roomRepo:     roomRepo,
		scheduleRepo: scheduleRepo,
		log:          log,
	}
}

func (s *ScheduleService) Create(ctx context.Context, roomID uuid.UUID, schedule *model.Schedule) error {
	if room, err := s.roomRepo.GetByID(ctx, roomID); err != nil {
		return err
	} else {
		if room == nil {
			return model.ErrRoomNotFound
		}
	}

	if sch, err := s.scheduleRepo.GetByRoomId(ctx, roomID); err != nil {
		return err
	} else {
		if sch != nil {
			s.log.Error("schedule already exists", "error", err)
			return model.ErrScheduleExists
		}
	}

	if err := s.scheduleRepo.Create(ctx, schedule); err != nil {
		return err
	}

	return nil
}
