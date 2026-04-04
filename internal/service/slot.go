package service

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-marlendd/internal/model"
)

type SlotService struct {
	slotRepo     SlotRepo
	scheduleRepo ScheduleRepo
	roomRepo     RoomRepo
	log          *slog.Logger
}

func NewSlotService(slotRepo SlotRepo, scheduleRepo ScheduleRepo, roomRepo RoomRepo, log *slog.Logger,
) *SlotService {
	return &SlotService{
		slotRepo:     slotRepo,
		scheduleRepo: scheduleRepo,
		roomRepo:     roomRepo,
		log:          log,
	}
}

func (s *SlotService) ListAvailable(ctx context.Context, roomID uuid.UUID, date time.Time) ([]model.Slot, error) {
	// 0. проверяем существование комнаты
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, model.ErrRoomNotFound
	}

	// 1. получаем расписание
	schedule, err := s.scheduleRepo.GetByRoomId(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if schedule == nil {
		s.log.Info("no schedule for roomID", "roomID", roomID)
		return []model.Slot{}, nil
	}

	// 2. проверяем дни недели
	day := int(date.Weekday())
	if day == 0 {
		day = 7 // в go sunday = 0, конвертируем
	}

	if !slices.Contains(schedule.DaysOfWeek, day) {
		return []model.Slot{}, nil
	}

	// 3. парсим startTime и endTime
	startT, err := time.Parse("15:04:05.999999", schedule.StartTime)
	if err != nil {
		startT, err = time.Parse("15:04", schedule.StartTime)
		if err != nil {
			s.log.Error("failed to parse start time", "error", err)
			return nil, err
		}
	}

	endT, err := time.Parse("15:04:05.999999", schedule.EndTime)
	if err != nil {
		endT, err = time.Parse("15:04", schedule.EndTime)
		if err != nil {
			s.log.Error("failed to parse end time", "error", err)
			return nil, err
		}
	}

	// 4. генерируем слоты
	slotStart := time.Date(date.Year(), date.Month(), date.Day(),
		startT.Hour(), startT.Minute(), 0, 0, time.UTC)
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(),
		endT.Hour(), endT.Minute(), 0, 0, time.UTC)

	var slots []model.Slot
	for slotStart.Before(dayEnd) {
		slotEnd := slotStart.Add(30 * time.Minute)
		if slotEnd.After(dayEnd) {
			break
		}
		slots = append(slots, model.Slot{
			ID:     uuid.New(),
			RoomID: roomID,
			Start:  slotStart,
			End:    slotEnd,
		})
		slotStart = slotEnd
	}

	if err := s.slotRepo.BulkUpsert(ctx, slots); err != nil {
		return nil, err
	}

	return s.slotRepo.ListAvailable(ctx, roomID, date)
}
