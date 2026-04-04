package service_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/internships-backend/test-backend-marlendd/internal/model"
	"github.com/internships-backend/test-backend-marlendd/internal/service"
)

var testLog = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

func futureSlot(id uuid.UUID) *model.Slot {
	return &model.Slot{
		ID:     id,
		RoomID: uuid.New(),
		Start:  time.Now().Add(time.Hour),
		End:    time.Now().Add(2 * time.Hour),
	}
}

func newBookingSvc(bookingRepo *mockBookingRepo, slotRepo *mockSlotRepo, conf *mockConferenceClient) *service.BookingService {
	return service.NewBookingService(bookingRepo, slotRepo, conf, testLog)
}

func TestBookingCreate_Success(t *testing.T) {
	slotID := uuid.New()
	userID := uuid.New()

	svc := newBookingSvc(
		&mockBookingRepo{
			createFn: func(_ context.Context, b *model.Booking) error {
				b.ID = uuid.New()
				b.CreatedAt = time.Now()
				return nil
			},
		},
		&mockSlotRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*model.Slot, error) {
				return futureSlot(id), nil
			},
		},
		nil,
	)

	booking, err := svc.Create(context.Background(), userID, slotID, false)
	require.NoError(t, err)
	assert.Equal(t, slotID, booking.SlotID)
	assert.Equal(t, userID, booking.UserID)
	assert.Equal(t, model.StatusActive, booking.Status)
}

func TestBookingCreate_SlotNotFound(t *testing.T) {
	svc := newBookingSvc(
		&mockBookingRepo{},
		&mockSlotRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Slot, error) {
				return nil, nil
			},
		},
		nil,
	)

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), false)
	assert.ErrorIs(t, err, model.ErrSlotNotFound)
}

func TestBookingCreate_SlotInPast(t *testing.T) {
	slotID := uuid.New()
	svc := newBookingSvc(
		&mockBookingRepo{},
		&mockSlotRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*model.Slot, error) {
				return &model.Slot{
					ID:    id,
					Start: time.Now().Add(-time.Hour),
					End:   time.Now().Add(-30 * time.Minute),
				}, nil
			},
		},
		nil,
	)

	_, err := svc.Create(context.Background(), uuid.New(), slotID, false)
	assert.ErrorIs(t, err, model.ErrSlotInPast)
}

func TestBookingCreate_AlreadyBooked(t *testing.T) {
	slotID := uuid.New()
	svc := newBookingSvc(
		&mockBookingRepo{
			createFn: func(_ context.Context, _ *model.Booking) error {
				return model.ErrSlotAlreadyBooked
			},
		},
		&mockSlotRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*model.Slot, error) {
				return futureSlot(id), nil
			},
		},
		nil,
	)

	_, err := svc.Create(context.Background(), uuid.New(), slotID, false)
	assert.ErrorIs(t, err, model.ErrSlotAlreadyBooked)
}

func TestBookingCreate_WithConferenceLink(t *testing.T) {
	slotID := uuid.New()
	userID := uuid.New()

	svc := newBookingSvc(
		&mockBookingRepo{
			createFn: func(_ context.Context, b *model.Booking) error {
				b.ID = uuid.New()
				return nil
			},
		},
		&mockSlotRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*model.Slot, error) {
				return futureSlot(id), nil
			},
		},
		&mockConferenceClient{
			createLinkFn: func(_ context.Context, _ uuid.UUID) (string, error) {
				return "https://meet.example.com/test", nil
			},
		},
	)

	booking, err := svc.Create(context.Background(), userID, slotID, true)
	require.NoError(t, err)
	require.NotNil(t, booking.ConferenceLink)
	assert.Equal(t, "https://meet.example.com/test", *booking.ConferenceLink)
}

func TestBookingCreate_ConferenceLinkFails_BookingStillCreated(t *testing.T) {
	slotID := uuid.New()
	userID := uuid.New()

	svc := newBookingSvc(
		&mockBookingRepo{
			createFn: func(_ context.Context, b *model.Booking) error {
				b.ID = uuid.New()
				return nil
			},
		},
		&mockSlotRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*model.Slot, error) {
				return futureSlot(id), nil
			},
		},
		&mockConferenceClient{
			createLinkFn: func(_ context.Context, _ uuid.UUID) (string, error) {
				return "", errors.New("conference service unavailable")
			},
		},
	)

	booking, err := svc.Create(context.Background(), userID, slotID, true)
	require.NoError(t, err) // бронь создана несмотря на сбой
	assert.Nil(t, booking.ConferenceLink)
}

func TestBookingCancel_Success(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()
	slotID := uuid.New()

	existing := &model.Booking{
		ID:     bookingID,
		SlotID: slotID,
		UserID: userID,
		Status: model.StatusActive,
	}

	svc := newBookingSvc(
		&mockBookingRepo{
			getByIdFn: func(_ context.Context, _ uuid.UUID) (*model.Booking, error) {
				return existing, nil
			},
			cancelFn: func(_ context.Context, _ uuid.UUID) error {
				return nil
			},
		},
		&mockSlotRepo{},
		nil,
	)

	booking, err := svc.Cancel(context.Background(), userID, bookingID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusCanceled, booking.Status)
}

func TestBookingCancel_Idempotent(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()

	existing := &model.Booking{
		ID:     bookingID,
		UserID: userID,
		Status: model.StatusCanceled,
	}

	svc := newBookingSvc(
		&mockBookingRepo{
			getByIdFn: func(_ context.Context, _ uuid.UUID) (*model.Booking, error) {
				return existing, nil
			},
		},
		&mockSlotRepo{},
		nil,
	)

	booking, err := svc.Cancel(context.Background(), userID, bookingID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusCanceled, booking.Status)
}

func TestBookingCancel_NotFound(t *testing.T) {
	svc := newBookingSvc(
		&mockBookingRepo{
			getByIdFn: func(_ context.Context, _ uuid.UUID) (*model.Booking, error) {
				return nil, nil
			},
		},
		&mockSlotRepo{},
		nil,
	)

	_, err := svc.Cancel(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, model.ErrBookingNotFound)
}

func TestBookingCancel_Forbidden(t *testing.T) {
	bookingID := uuid.New()
	ownerID := uuid.New()
	otherID := uuid.New()

	svc := newBookingSvc(
		&mockBookingRepo{
			getByIdFn: func(_ context.Context, _ uuid.UUID) (*model.Booking, error) {
				return &model.Booking{
					ID:     bookingID,
					UserID: ownerID,
					Status: model.StatusActive,
				}, nil
			},
		},
		&mockSlotRepo{},
		nil,
	)

	_, err := svc.Cancel(context.Background(), otherID, bookingID)
	assert.ErrorIs(t, err, model.ErrForbidden)
}

func TestBookingCreate_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newBookingSvc(
		&mockBookingRepo{
			createFn: func(_ context.Context, _ *model.Booking) error {
				return repoErr
			},
		},
		&mockSlotRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*model.Slot, error) {
				return futureSlot(id), nil
			},
		},
		nil,
	)

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), false)
	assert.ErrorIs(t, err, repoErr)
}
