package repository

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-marlendd/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleRepository struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewScheduleRepository(db *pgxpool.Pool, log *slog.Logger) *ScheduleRepository {
	return &ScheduleRepository{
		db:  db,
		log: log,
	}
}

func (s *ScheduleRepository) Create(ctx context.Context, schedule *model.Schedule) error {
	query := `INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time)
			VALUES ($1, $2, $3, $4, $5)`

	id := uuid.New()
	if _, err := s.db.Exec(ctx, query,
		id, schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime,
	); err != nil {
		s.log.Error("failed to execute query", "error", err)
		return err
	}

	schedule.ID = id

	return nil
}

func (s *ScheduleRepository) GetByRoomId(ctx context.Context, roomID uuid.UUID) (*model.Schedule, error) {
	query := `SELECT id, room_id, days_of_week, start_time, end_time 
			FROM schedules
			WHERE room_id = $1`

	var schedule model.Schedule
	row := s.db.QueryRow(ctx, query, roomID)

	if err := row.Scan(
		&schedule.ID, &schedule.RoomID, &schedule.DaysOfWeek, &schedule.StartTime, &schedule.EndTime,
	); err != nil {
		if err == pgx.ErrNoRows {
			s.log.Info("no such schedule for room", "roomID", roomID)
			return nil, nil
		}
		s.log.Error("failed to scan db response", "error", err)
		return nil, err
	}
	return &schedule, nil
}
