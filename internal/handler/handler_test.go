package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"room-booking/internal/model"
)

// --- Mock services ---

type mockAuthSvc struct {
	secret string
}

func (m *mockAuthSvc) DummyLogin(role string) (string, error) {
	if role != "admin" && role != "user" {
		return "", fmt.Errorf("invalid role")
	}
	return "mock-token-" + role, nil
}

func (m *mockAuthSvc) ParseToken(token string) (uuid.UUID, string, error) {
	switch token {
	case "mock-token-admin":
		return uuid.MustParse("00000000-0000-0000-0000-000000000001"), "admin", nil
	case "mock-token-user":
		return uuid.MustParse("00000000-0000-0000-0000-000000000002"), "user", nil
	default:
		return uuid.Nil, "", fmt.Errorf("invalid token")
	}
}

type mockRoomSvc struct {
	rooms []model.Room
}

func (m *mockRoomSvc) Create(_ context.Context, name string, desc *string, cap *int) (*model.Room, error) {
	if name == "" {
		return nil, model.NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "name is required")
	}
	now := time.Now()
	room := &model.Room{ID: uuid.New(), Name: name, Description: desc, Capacity: cap, CreatedAt: &now}
	return room, nil
}

func (m *mockRoomSvc) List(_ context.Context) ([]model.Room, error) {
	return m.rooms, nil
}

type mockScheduleSvc struct{}

func (m *mockScheduleSvc) Create(_ context.Context, roomID uuid.UUID, daysOfWeek []int, startTime, endTime string) (*model.Schedule, error) {
	return &model.Schedule{
		ID: uuid.New(), RoomID: roomID, DaysOfWeek: daysOfWeek, StartTime: startTime, EndTime: endTime,
	}, nil
}

type mockSlotSvc struct{}

func (m *mockSlotSvc) ListAvailable(_ context.Context, _ uuid.UUID, _ time.Time) ([]model.Slot, error) {
	return []model.Slot{
		{ID: uuid.New(), RoomID: uuid.New(), Start: time.Now().Add(time.Hour), End: time.Now().Add(90 * time.Minute)},
	}, nil
}

type mockBookingSvc struct {
	bookings []model.Booking
}

func (m *mockBookingSvc) Create(_ context.Context, userID, slotID uuid.UUID, _ bool) (*model.Booking, error) {
	now := time.Now()
	return &model.Booking{ID: uuid.New(), SlotID: slotID, UserID: userID, Status: "active", CreatedAt: &now}, nil
}

func (m *mockBookingSvc) ListAll(_ context.Context, page, pageSize int) ([]model.Booking, *model.Pagination, error) {
	return m.bookings, &model.Pagination{Page: page, PageSize: pageSize, Total: len(m.bookings)}, nil
}

func (m *mockBookingSvc) ListMy(_ context.Context, _ uuid.UUID) ([]model.Booking, error) {
	return m.bookings, nil
}

func (m *mockBookingSvc) Cancel(_ context.Context, _, bookingID uuid.UUID) (*model.Booking, error) {
	now := time.Now()
	return &model.Booking{ID: bookingID, Status: "cancelled", CreatedAt: &now, SlotID: uuid.New(), UserID: uuid.MustParse("00000000-0000-0000-0000-000000000002")}, nil
}

func newTestHandler() *Handler {
	return New(
		&mockAuthSvc{},
		&mockRoomSvc{rooms: make([]model.Room, 0)},
		&mockScheduleSvc{},
		&mockSlotSvc{},
		&mockBookingSvc{bookings: make([]model.Booking, 0)},
	)
}

// --- Tests ---

func TestInfo(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	req := httptest.NewRequest("GET", "/_info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestInfoRoot(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestDummyLogin_ValidAdmin(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	body, _ := json.Marshal(map[string]string{"role": "admin"})
	req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["token"] == "" {
		t.Error("expected non-empty token")
	}
}

func TestDummyLogin_ValidUser(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	body, _ := json.Marshal(map[string]string{"role": "user"})
	req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestDummyLogin_InvalidRole(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	body, _ := json.Marshal(map[string]string{"role": "superadmin"})
	req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDummyLogin_InvalidBody(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestListRooms_Unauthorized(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	req := httptest.NewRequest("GET", "/rooms/list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestListRooms_Success(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	req := httptest.NewRequest("GET", "/rooms/list", nil)
	req.Header.Set("Authorization", "Bearer mock-token-user")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCreateRoom_AdminOnly(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	body, _ := json.Marshal(map[string]string{"name": "Test Room"})
	req := httptest.NewRequest("POST", "/rooms/create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer mock-token-user")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for user creating room, got %d", w.Code)
	}
}

func TestCreateRoom_AdminSuccess(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	body, _ := json.Marshal(map[string]string{"name": "Test Room"})
	req := httptest.NewRequest("POST", "/rooms/create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer mock-token-admin")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestCreateRoom_EmptyName(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	body, _ := json.Marshal(map[string]string{"name": ""})
	req := httptest.NewRequest("POST", "/rooms/create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer mock-token-admin")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateBooking_AdminForbidden(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	body, _ := json.Marshal(map[string]string{"slotId": uuid.New().String()})
	req := httptest.NewRequest("POST", "/bookings/create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer mock-token-admin")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for admin creating booking, got %d", w.Code)
	}
}

func TestCreateBooking_UserSuccess(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	body, _ := json.Marshal(map[string]string{"slotId": uuid.New().String()})
	req := httptest.NewRequest("POST", "/bookings/create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer mock-token-user")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestListBookings_AdminOnly(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	req := httptest.NewRequest("GET", "/bookings/list", nil)
	req.Header.Set("Authorization", "Bearer mock-token-user")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for user listing all bookings, got %d", w.Code)
	}
}

func TestListBookings_AdminSuccess(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	req := httptest.NewRequest("GET", "/bookings/list", nil)
	req.Header.Set("Authorization", "Bearer mock-token-admin")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestListMyBookings_UserOnly(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	req := httptest.NewRequest("GET", "/bookings/my", nil)
	req.Header.Set("Authorization", "Bearer mock-token-admin")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for admin listing my bookings, got %d", w.Code)
	}
}

func TestListMyBookings_UserSuccess(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	req := httptest.NewRequest("GET", "/bookings/my", nil)
	req.Header.Set("Authorization", "Bearer mock-token-user")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCancelBooking_UserSuccess(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	bookingID := uuid.New()
	req := httptest.NewRequest("POST", "/bookings/"+bookingID.String()+"/cancel", nil)
	req.Header.Set("Authorization", "Bearer mock-token-user")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCancelBooking_AdminForbidden(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	bookingID := uuid.New()
	req := httptest.NewRequest("POST", "/bookings/"+bookingID.String()+"/cancel", nil)
	req.Header.Set("Authorization", "Bearer mock-token-admin")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestListSlots_MissingDate(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	roomID := uuid.New()
	req := httptest.NewRequest("GET", "/rooms/"+roomID.String()+"/slots/list", nil)
	req.Header.Set("Authorization", "Bearer mock-token-user")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing date, got %d", w.Code)
	}
}

func TestListSlots_InvalidDate(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	roomID := uuid.New()
	req := httptest.NewRequest("GET", "/rooms/"+roomID.String()+"/slots/list?date=not-a-date", nil)
	req.Header.Set("Authorization", "Bearer mock-token-user")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid date, got %d", w.Code)
	}
}

func TestListSlots_ValidDate(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	roomID := uuid.New()
	req := httptest.NewRequest("GET", "/rooms/"+roomID.String()+"/slots/list?date=2026-04-01", nil)
	req.Header.Set("Authorization", "Bearer mock-token-user")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCreateSchedule_AdminSuccess(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	roomID := uuid.New()
	body, _ := json.Marshal(map[string]interface{}{
		"daysOfWeek": []int{1, 2, 3, 4, 5},
		"startTime":  "09:00",
		"endTime":    "18:00",
	})
	req := httptest.NewRequest("POST", "/rooms/"+roomID.String()+"/schedule/create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer mock-token-admin")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestCreateSchedule_UserForbidden(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	roomID := uuid.New()
	body, _ := json.Marshal(map[string]interface{}{
		"daysOfWeek": []int{1, 2, 3},
		"startTime":  "09:00",
		"endTime":    "18:00",
	})
	req := httptest.NewRequest("POST", "/rooms/"+roomID.String()+"/schedule/create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer mock-token-user")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestInvalidToken(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	req := httptest.NewRequest("GET", "/rooms/list", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestMissingAuthHeader(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	req := httptest.NewRequest("GET", "/rooms/list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCreateBooking_InvalidSlotID(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	body, _ := json.Marshal(map[string]string{"slotId": "not-a-uuid"})
	req := httptest.NewRequest("POST", "/bookings/create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer mock-token-user")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCancelBooking_InvalidBookingID(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	req := httptest.NewRequest("POST", "/bookings/not-a-uuid/cancel", nil)
	req.Header.Set("Authorization", "Bearer mock-token-user")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestListBookings_Pagination(t *testing.T) {
	h := newTestHandler()
	router := h.Router()

	req := httptest.NewRequest("GET", "/bookings/list?page=2&pageSize=10", nil)
	req.Header.Set("Authorization", "Bearer mock-token-admin")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	pag, ok := resp["pagination"].(map[string]interface{})
	if !ok {
		t.Fatal("expected pagination in response")
	}
	if int(pag["page"].(float64)) != 2 {
		t.Errorf("expected page=2, got %v", pag["page"])
	}
	if int(pag["pageSize"].(float64)) != 10 {
		t.Errorf("expected pageSize=10, got %v", pag["pageSize"])
	}
}
