package service

import (
	"context"
	"log/slog"

	"github.com/internships-backend/test-backend-marlendd/internal/model"
)

type RoomService struct {
	roomRepo RoomRepo
	log      *slog.Logger
}

func NewRoomService(roomRepo RoomRepo, log *slog.Logger) *RoomService {
	return &RoomService{
		roomRepo: roomRepo,
		log:      log,
	}
}

func (r *RoomService) Create(ctx context.Context, room *model.Room) error {
	return r.roomRepo.Create(ctx, room)
}

func (r *RoomService) List(ctx context.Context) ([]model.Room, error) {
	return r.roomRepo.List(ctx)
}
