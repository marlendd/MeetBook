package model

import "errors"

var (
	ErrSlotNotFound      = errors.New("slot not found")           // 404
	ErrSlotInPast        = errors.New("slot is in the past")      // 400
	ErrSlotAlreadyBooked = errors.New("slot is already booked")   // 409
	ErrBookingNotFound   = errors.New("booking not found")        // 404
	ErrForbidden         = errors.New("forbidden")                // 403
	ErrRoomNotFound      = errors.New("room not found")           // 404
	ErrScheduleExists    = errors.New("schedule already exitsts") // 409
	ErrInvalidRole       = errors.New("invalid role")
)
