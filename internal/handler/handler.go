package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"room-booking/internal/middleware"
	"room-booking/internal/model"
)

type AuthSvc interface {
	DummyLogin(role string) (string, error)
	ParseToken(token string) (uuid.UUID, string, error)
}

type RoomSvc interface {
	Create(ctx context.Context, name string, desc *string, cap *int) (*model.Room, error)
	List(ctx context.Context) ([]model.Room, error)
}

type ScheduleSvc interface {
	Create(ctx context.Context, roomID uuid.UUID, daysOfWeek []int, startTime, endTime string) (*model.Schedule, error)
}

type SlotSvc interface {
	ListAvailable(ctx context.Context, roomID uuid.UUID, date time.Time) ([]model.Slot, error)
}

type BookingSvc interface {
	Create(ctx context.Context, userID, slotID uuid.UUID, createConferenceLink bool) (*model.Booking, error)
	ListAll(ctx context.Context, page, pageSize int) ([]model.Booking, *model.Pagination, error)
	ListMy(ctx context.Context, userID uuid.UUID) ([]model.Booking, error)
	Cancel(ctx context.Context, userID, bookingID uuid.UUID) (*model.Booking, error)
}

type Handler struct {
	auth      AuthSvc
	rooms     RoomSvc
	schedules ScheduleSvc
	slots     SlotSvc
	bookings  BookingSvc
}

func New(auth AuthSvc, rooms RoomSvc, schedules ScheduleSvc, slots SlotSvc, bookings BookingSvc) *Handler {
	return &Handler{
		auth:      auth,
		rooms:     rooms,
		schedules: schedules,
		slots:     slots,
		bookings:  bookings,
	}
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)

	r.Get("/", h.Info)
	r.Get("/_info", h.Info)
	r.Post("/dummyLogin", h.DummyLogin)

	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)

		r.Get("/rooms/list", h.ListRooms)
		r.Post("/rooms/create", h.requireRole("admin", h.CreateRoom))
		r.Post("/rooms/{roomId}/schedule/create", h.requireRole("admin", h.CreateSchedule))
		r.Get("/rooms/{roomId}/slots/list", h.ListSlots)

		r.Post("/bookings/create", h.requireRole("user", h.CreateBooking))
		r.Get("/bookings/list", h.requireRole("admin", h.ListBookings))
		r.Get("/bookings/my", h.requireRole("user", h.ListMyBookings))
		r.Post("/bookings/{bookingId}/cancel", h.requireRole("user", h.CancelBooking))
	})

	return r
}

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid authorization header")
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, role, err := h.auth.ParseToken(tokenStr)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
			return
		}
		ctx := middleware.SetClaims(r.Context(), &middleware.Claims{UserID: userID, Role: role})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) requireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.GetClaims(r.Context())
		if claims == nil || claims.Role != role {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied: requires role "+role)
			return
		}
		next(w, r)
	}
}
