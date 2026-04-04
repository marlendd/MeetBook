package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/internships-backend/test-backend-marlendd/internal/config"
	"github.com/internships-backend/test-backend-marlendd/internal/db"
	"github.com/internships-backend/test-backend-marlendd/internal/handler"
	"github.com/internships-backend/test-backend-marlendd/internal/middleware"
	"github.com/internships-backend/test-backend-marlendd/internal/model"
	"github.com/internships-backend/test-backend-marlendd/internal/repository"
	"github.com/internships-backend/test-backend-marlendd/internal/service"
)

func main() {
	// config
	cfg := config.MustLoad()

	//logger
	log := setupLogger(cfg.LogLevel)

	if err := run(cfg, log); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting server")
	log.Debug("debug messages are enabled")

	// БД
	dbCtx := context.Background()
	pool, err := db.Connect(dbCtx, cfg.PostgresDSN, log)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}

	// миграции
	if err := db.RunMigrations(cfg.PostgresDSN, log); err != nil {
		return fmt.Errorf("failed to run db migrations: %w", err)
	}

	// репозитории
	roomRepo := repository.NewRoomRepository(pool, log)
	scheduleRepo := repository.NewScheduleRepository(pool, log)
	slotRepo := repository.NewSlotRepository(pool, log)
	bookingRepo := repository.NewBookingRepository(pool, log)

	// сервисы
	authService := service.NewAuthService(cfg.JWTSecret)
	roomService := service.NewRoomService(roomRepo, log)
	scheduleService := service.NewScheduleService(roomRepo, scheduleRepo, log)
	slotService := service.NewSlotService(slotRepo, scheduleRepo, roomRepo, log)
	bookingService := service.NewBookingService(bookingRepo, slotRepo, log)

	// хэндлеры
	authHandler := handler.NewAuthHandler(authService, log)
	roomHandler := handler.NewRoomHandler(roomService, log)
	scheduleHandler := handler.NewScheduleHandler(scheduleService, log)
	slotHandler := handler.NewSlotsHandler(slotService, log)
	bookingHandler := handler.NewBookingHandler(bookingService, log)

	// роутер
	mux := http.NewServeMux()

	auth := middleware.Auth(cfg.JWTSecret)
	adminOnly := middleware.RequireRole(model.RoleAdmin)
	userOnly := middleware.RequireRole(model.RoleUser)

	mux.HandleFunc("GET /_info", handler.InfoHandler)

	mux.HandleFunc("POST /dummyLogin", authHandler.DummyLogin)

	// rooms
	mux.Handle("POST /rooms/create", auth(adminOnly(http.HandlerFunc(roomHandler.Create))))
	mux.Handle("GET /rooms/list", auth(http.HandlerFunc(roomHandler.List)))

	mux.Handle("POST /rooms/{roomId}/schedule/create", auth(adminOnly(http.HandlerFunc(scheduleHandler.Create))))
	mux.Handle("GET /rooms/{roomId}/slots/list", auth(http.HandlerFunc(slotHandler.ListAvailable)))

	// bookings
	mux.Handle("POST /bookings/create", auth(userOnly(http.HandlerFunc(bookingHandler.Create))))
	mux.Handle("GET /bookings/my", auth(userOnly(http.HandlerFunc(bookingHandler.ListByUser))))
	mux.Handle("POST /bookings/{bookingId}/cancel", auth(userOnly(http.HandlerFunc(bookingHandler.Cancel))))
	mux.Handle("GET /bookings/list", auth(adminOnly(http.HandlerFunc(bookingHandler.ListAll))))

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("starting server", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}

func setupLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
