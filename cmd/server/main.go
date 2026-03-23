package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"room-booking/internal/config"
	"room-booking/internal/handler"
	"room-booking/internal/service"
	"room-booking/internal/store"
)

func main() {
	cfg := config.Load()

	db, err := connectWithRetry(cfg.DatabaseURL(), 30)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	authSvc := service.NewAuthService(cfg.JWTSecret)
	roomSvc := service.NewRoomService(db)
	scheduleSvc := service.NewScheduleService(db)
	slotSvc := service.NewSlotService(db)
	bookingSvc := service.NewBookingService(db)

	h := handler.New(authSvc, roomSvc, scheduleSvc, slotSvc, bookingSvc)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, h.Router()); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func connectWithRetry(url string, maxRetries int) (*store.Store, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		s, err := store.New(url)
		if err == nil {
			log.Println("Connected to database")
			return s, nil
		}
		lastErr = err
		log.Printf("Database connection attempt %d/%d: %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}
