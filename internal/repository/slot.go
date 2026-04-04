package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/internships-backend/test-backend-marlendd/internal/model"
)

type SlotRepository struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewSlotRepository(db *pgxpool.Pool, log *slog.Logger) *SlotRepository {
	return &SlotRepository{
		db:  db,
		log: log,
	}
}

func (s *SlotRepository) BulkUpsert(ctx context.Context, slots []model.Slot) error {
	if len(slots) == 0 {
		s.log.Info("empty slots to upsert")
		return nil
	}

	var sb strings.Builder

	sb.WriteString(`INSERT INTO slots(id, room_id, start, "end") VALUES`)
	args := make([]any, 0, len(slots)*4)

	for i, slot := range slots {
		if i > 0 {
			sb.WriteString(",")
		}
		base := i * 4
		fmt.Fprintf(&sb, "($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4)
		args = append(args, slot.ID, slot.RoomID, slot.Start, slot.End)
	}
	sb.WriteString(` ON CONFLICT (room_id, start) DO NOTHING`)

	query := sb.String()

	if _, err := s.db.Exec(ctx, query, args...); err != nil {
		s.log.Error("failed to bulf upsert slots", "error", err)
		return err
	}

	return nil
}

func (s *SlotRepository) ListAvailable(
	ctx context.Context, roomID uuid.UUID, date time.Time,
) ([]model.Slot, error) {
	dayStart := time.Date(
		date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC,
	)
	dayEnd := dayStart.Add(24 * time.Hour)

	query := `
				SELECT s.id, s.room_id, s.start, s."end"
				FROM slots s
				WHERE s.room_id = $1
				AND s.start >= $2
				AND s.start < $3
				AND NOT EXISTS (
					SELECT 1 FROM bookings b
					WHERE b.slot_id = s.id AND b.status = 'active'
				)
				ORDER BY s.start`

	rows, err := s.db.Query(ctx, query, roomID, dayStart, dayEnd)
	if err != nil {
		s.log.Error("failed to query available slots", "error", err)
		return nil, err
	}
	defer rows.Close()

	var slots []model.Slot
	for rows.Next() {
		var slot model.Slot
		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End); err != nil {
			s.log.Error("failed to scan slot", "error", err)
			return nil, err
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return slots, nil
}

func (s *SlotRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Slot, error) {
	query := `SELECT id, room_id, start, "end" FROM slots WHERE id = $1`

	var slot model.Slot
	err := s.db.QueryRow(ctx, query, id).Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		s.log.Error("failed to get slot by id", "error", err)
		return nil, err
	}
	return &slot, nil
}
