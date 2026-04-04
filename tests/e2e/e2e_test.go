package e2e_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-marlendd/internal/db"
	"github.com/internships-backend/test-backend-marlendd/internal/model"
	"github.com/internships-backend/test-backend-marlendd/internal/repository"
	"github.com/internships-backend/test-backend-marlendd/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testLog = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

func setupDB(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	pgc, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("booking"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pgc.Terminate(ctx) })

	dsn, err := pgc.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	err = db.RunMigrations(dsn, testLog)
	require.NoError(t, err)

	// вставляем фиксированных пользователей (требование FK в таблице bookings)
	pool, err := db.Connect(ctx, dsn, testLog)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, role) VALUES
		('00000000-0000-0000-0000-000000000001', 'admin@example.com', 'admin'),
		('00000000-0000-0000-0000-000000000002', 'user@example.com', 'user')
		ON CONFLICT DO NOTHING
	`)
	require.NoError(t, err)

	return dsn
}

// TestE2E_CreateRoomScheduleBooking проверяет сценарий:
// создание переговорки → создание расписания → получение слотов → создание брони
func TestE2E_CreateRoomScheduleBooking(t *testing.T) {
	dsn := setupDB(t)
	ctx := context.Background()

	pool, err := db.Connect(ctx, dsn, testLog)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	roomRepo := repository.NewRoomRepository(pool, testLog)
	scheduleRepo := repository.NewScheduleRepository(pool, testLog)
	slotRepo := repository.NewSlotRepository(pool, testLog)
	bookingRepo := repository.NewBookingRepository(pool, testLog)

	roomSvc := service.NewRoomService(roomRepo, testLog)
	scheduleSvc := service.NewScheduleService(roomRepo, scheduleRepo, testLog)
	slotSvc := service.NewSlotService(slotRepo, scheduleRepo, roomRepo, testLog)
	bookingSvc := service.NewBookingService(bookingRepo, slotRepo, testLog)

	// 1. Создаём переговорку
	room := &model.Room{Name: "Conference Room A"}
	err = roomSvc.Create(ctx, room)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, room.ID)

	// 2. Создаём расписание (все дни, 09:00-18:00)
	tomorrow := time.Now().UTC().AddDate(0, 0, 1)
	dayOfWeek := int(tomorrow.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7
	}

	schedule := &model.Schedule{
		RoomID:     room.ID,
		DaysOfWeek: []int{1, 2, 3, 4, 5, 6, 7},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}
	err = scheduleSvc.Create(ctx, room.ID, schedule)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, schedule.ID)

	// 3. Получаем доступные слоты
	slots, err := slotSvc.ListAvailable(ctx, room.ID, tomorrow)
	require.NoError(t, err)
	assert.NotEmpty(t, slots)
	// 9:00-18:00 = 18 слотов по 30 минут
	assert.Len(t, slots, 18)

	// 4. Создаём бронь
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	booking, err := bookingSvc.Create(ctx, userID, slots[0].ID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusActive, booking.Status)
	assert.Equal(t, slots[0].ID, booking.SlotID)
	assert.Equal(t, userID, booking.UserID)
}

// TestE2E_CancelBooking проверяет сценарий: создание брони → отмена → повторная отмена (идемпотентность)
func TestE2E_CancelBooking(t *testing.T) {
	dsn := setupDB(t)
	ctx := context.Background()

	pool, err := db.Connect(ctx, dsn, testLog)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	roomRepo := repository.NewRoomRepository(pool, testLog)
	scheduleRepo := repository.NewScheduleRepository(pool, testLog)
	slotRepo := repository.NewSlotRepository(pool, testLog)
	bookingRepo := repository.NewBookingRepository(pool, testLog)

	roomSvc := service.NewRoomService(roomRepo, testLog)
	scheduleSvc := service.NewScheduleService(roomRepo, scheduleRepo, testLog)
	slotSvc := service.NewSlotService(slotRepo, scheduleRepo, roomRepo, testLog)
	bookingSvc := service.NewBookingService(bookingRepo, slotRepo, testLog)

	// Setup: room + schedule + slot
	room := &model.Room{Name: "Room B"}
	require.NoError(t, roomSvc.Create(ctx, room))

	tomorrow := time.Now().UTC().AddDate(0, 0, 1)
	require.NoError(t, scheduleSvc.Create(ctx, room.ID, &model.Schedule{
		RoomID:     room.ID,
		DaysOfWeek: []int{1, 2, 3, 4, 5, 6, 7},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}))

	slots, err := slotSvc.ListAvailable(ctx, room.ID, tomorrow)
	require.NoError(t, err)
	require.NotEmpty(t, slots)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	// Создаём бронь
	booking, err := bookingSvc.Create(ctx, userID, slots[0].ID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusActive, booking.Status)

	// Отменяем бронь
	cancelled, err := bookingSvc.Cancel(ctx, userID, booking.ID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusCancelled, cancelled.Status)

	// Повторная отмена — идемпотентна, ошибки нет
	cancelled2, err := bookingSvc.Cancel(ctx, userID, booking.ID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusCancelled, cancelled2.Status)

	// Слот должен снова стать доступным
	availableSlots, err := slotSvc.ListAvailable(ctx, room.ID, tomorrow)
	require.NoError(t, err)
	var found bool
	for _, s := range availableSlots {
		if s.ID == slots[0].ID {
			found = true
			break
		}
	}
	assert.True(t, found, "отменённый слот должен снова быть доступен")
}
