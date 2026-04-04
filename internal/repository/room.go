package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-marlendd/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewRoomRepository(db *pgxpool.Pool, log *slog.Logger) *RoomRepository {
	return &RoomRepository{
		db:  db,
		log: log,
	}
}

func (r *RoomRepository) Create(ctx context.Context, room *model.Room) error {
	query := `INSERT INTO rooms (id, name, description, capacity, created_at)
			VALUES ($1, $2, $3, $4, $5)`

	id := uuid.New()
	now := time.Now().UTC()

	_, err := r.db.Exec(ctx, query,
		id, room.Name, room.Description, room.Capacity, now,
	)
	if err != nil {
		r.log.Error("failed to insert data", "error", err)
		return err
	}

	room.ID = id
	room.CreatedAt = now

	return nil
}

func (r *RoomRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Room, error) {
	query := `SELECT id, name, description, capacity, created_at FROM rooms WHERE id = $1`

	var room model.Room
	err := r.db.QueryRow(ctx, query, id).Scan(
		&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.log.Error("failed to get room by id", "error", err)
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) List(ctx context.Context) ([]model.Room, error) {
	query := `SELECT id, name, description, capacity, created_at 
			FROM rooms`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	var res []model.Room

	for rows.Next() {
		var room model.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt); err != nil {
			r.log.Error("failed to scan db response", "error", err)
			return nil, err
		}
		res = append(res, room)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("failed to get db response", "error", err)
		return nil, err
	}

	return res, nil
}
