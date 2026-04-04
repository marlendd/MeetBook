package model

import "errors"

var (
	ErrSlotNotFound       = errors.New("slot not found")
	ErrSlotInPast         = errors.New("slot is in the past")
	ErrSlotAlreadyBooked  = errors.New("slot is already booked")
	ErrBookingNotFound    = errors.New("booking not found")
	ErrForbidden          = errors.New("forbidden")
	ErrRoomNotFound       = errors.New("room not found")
	ErrScheduleExists     = errors.New("schedule already exitsts")
	ErrInvalidRole        = errors.New("invalid role")
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
