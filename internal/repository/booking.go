package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/internships-backend/test-backend-marlendd/internal/model"
)

type BookingRepository struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewBookingRepository(db *pgxpool.Pool, log *slog.Logger) *BookingRepository {
	return &BookingRepository{
		db:  db,
		log: log,
	}
}

func (b *BookingRepository) Create(ctx context.Context, booking *model.Booking) error {
	query := `INSERT INTO bookings (id, slot_id, user_id, status, conference_link, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)`

	id := uuid.New()
	now := time.Now().UTC()

	_, err := b.db.Exec(ctx, query,
		id, booking.SlotID, booking.UserID, booking.Status, booking.ConferenceLink, now,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.ErrSlotAlreadyBooked
		}
		b.log.Error("failed to insert data", "error", err)
		return err
	}

	booking.ID = id
	booking.CreatedAt = now

	return nil
}

func (b *BookingRepository) GetById(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	query := `SELECT id, slot_id, user_id, status, conference_link, created_at
			FROM bookings
			WHERE id = $1`

	var booking model.Booking
	row := b.db.QueryRow(ctx, query, id)

	if err := row.Scan(
		&booking.ID, &booking.SlotID, &booking.UserID, &booking.Status, &booking.ConferenceLink, &booking.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			b.log.Info("no such booking", "ID", id)
			return nil, nil
		}
		b.log.Error("failed to scan db response", "error", err)
		return nil, err
	}
	return &booking, nil
}

func (b *BookingRepository) Cancel(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE bookings SET status = 'canceled' WHERE id = $1`

	if _, err := b.db.Exec(ctx, query, id); err != nil {
		b.log.Error("failed to execute query", "error", err)
		return err
	}

	return nil
}

func (b *BookingRepository) UpdateConferenceLink(ctx context.Context, id uuid.UUID, link string) error {
	query := `UPDATE bookings SET conference_link = $1 WHERE id = $2`
	if _, err := b.db.Exec(ctx, query, link, id); err != nil {
		b.log.Error("failed to update conference link", "error", err)
		return err
	}
	return nil
}

func (b *BookingRepository) ListAll(ctx context.Context, page, pageSize int) ([]model.Booking, int, error) {
	var total int
	err := b.db.QueryRow(ctx, `SELECT COUNT(*) FROM bookings`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, slot_id, user_id, status, conference_link, created_at
			FROM bookings
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`

	offset := (page - 1) * pageSize
	rows, err := b.db.Query(ctx, query, pageSize, offset)
	if err != nil {
		b.log.Error("failed to execute query", "error", err)
		return nil, 0, err
	}
	defer rows.Close()

	var bookings []model.Booking
	for rows.Next() {
		var booking model.Booking
		if err := rows.Scan(&booking.ID, &booking.SlotID, &booking.UserID, &booking.Status, &booking.ConferenceLink, &booking.CreatedAt); err != nil {
			b.log.Error("failed to scan booking", "error", err)
			return nil, 0, err
		}
		bookings = append(bookings, booking)
	}
	if err := rows.Err(); err != nil {
		b.log.Error("failed to process scan", "error", err)
		return nil, 0, err
	}

	return bookings, total, nil
}

func (b *BookingRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Booking, error) {
	query := `SELECT b.id, b.slot_id, b.user_id, b.status, b.conference_link, b.created_at
			FROM bookings b
			JOIN slots s ON s.id = b.slot_id
			WHERE b.user_id = $1
			AND s.start >= NOW()
			ORDER BY s.start`

	rows, err := b.db.Query(ctx, query, userID)
	if err != nil {
		b.log.Error("failed to execute query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var bookings []model.Booking
	for rows.Next() {
		var booking model.Booking
		if err := rows.Scan(&booking.ID, &booking.SlotID, &booking.UserID, &booking.Status, &booking.ConferenceLink, &booking.CreatedAt); err != nil {
			b.log.Error("failed to scan booking", "error", err)
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		b.log.Error("failed to process scan", "error", err)
		return nil, err
	}

	return bookings, nil
}
