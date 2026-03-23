package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var baseURL string

func TestMain(m *testing.M) {
	baseURL = os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	os.Exit(m.Run())
}

func skipIfNoServer(t *testing.T) {
	t.Helper()
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("Skipping integration test; set INTEGRATION=true and ensure server is running")
	}
}

func TestIntegration_BookingFlow(t *testing.T) {
	skipIfNoServer(t)

	adminToken := getDummyToken(t, "admin")
	userToken := getDummyToken(t, "user")

	// 1. Create a room
	roomID := createRoom(t, adminToken, "Integration Test Room")

	// 2. Create a schedule (all days, 09:00-18:00)
	createSchedule(t, adminToken, roomID, []int{1, 2, 3, 4, 5, 6, 7}, "09:00", "18:00")

	// 3. List slots for a future date
	futureDate := findFutureDate()
	slots := listSlots(t, userToken, roomID, futureDate)
	if len(slots) == 0 {
		t.Fatal("expected at least one available slot")
	}

	slotID := slots[0]["id"].(string)

	// 4. Book a slot
	bookingID := createBooking(t, userToken, slotID)

	// 5. Verify slot is no longer available
	slotsAfter := listSlots(t, userToken, roomID, futureDate)
	for _, s := range slotsAfter {
		if s["id"].(string) == slotID {
			t.Error("booked slot should not appear in available slots")
		}
	}

	// 6. Verify booking appears in my bookings
	myBookings := listMyBookings(t, userToken)
	found := false
	for _, b := range myBookings {
		if b["id"].(string) == bookingID {
			found = true
			break
		}
	}
	if !found {
		t.Error("booking should appear in my bookings")
	}

	t.Logf("Booking flow test passed: room=%s, slot=%s, booking=%s", roomID, slotID, bookingID)
}

func TestIntegration_CancelBooking(t *testing.T) {
	skipIfNoServer(t)

	adminToken := getDummyToken(t, "admin")
	userToken := getDummyToken(t, "user")

	roomID := createRoom(t, adminToken, "Cancel Test Room")
	createSchedule(t, adminToken, roomID, []int{1, 2, 3, 4, 5, 6, 7}, "09:00", "18:00")

	futureDate := findFutureDate()
	slots := listSlots(t, userToken, roomID, futureDate)
	if len(slots) == 0 {
		t.Fatal("expected at least one available slot")
	}
	slotID := slots[0]["id"].(string)

	bookingID := createBooking(t, userToken, slotID)

	// Cancel booking
	cancelled := cancelBooking(t, userToken, bookingID)
	if cancelled["status"].(string) != "cancelled" {
		t.Errorf("expected status 'cancelled', got %s", cancelled["status"])
	}

	// Idempotent cancel
	cancelledAgain := cancelBooking(t, userToken, bookingID)
	if cancelledAgain["status"].(string) != "cancelled" {
		t.Errorf("expected status 'cancelled' on second cancel, got %s", cancelledAgain["status"])
	}

	// Slot should be available again
	slotsAfterCancel := listSlots(t, userToken, roomID, futureDate)
	found := false
	for _, s := range slotsAfterCancel {
		if s["id"].(string) == slotID {
			found = true
			break
		}
	}
	if !found {
		t.Error("cancelled slot should reappear as available")
	}

	t.Logf("Cancel booking test passed: booking=%s", bookingID)
}

// --- Helpers ---

func findFutureDate() string {
	d := time.Now().UTC().AddDate(0, 0, 1)
	return d.Format("2006-01-02")
}

func getDummyToken(t *testing.T, role string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"role": role})
	resp := doRequest(t, "POST", baseURL+"/dummyLogin", nil, body)
	assertStatus(t, resp, http.StatusOK)

	var result map[string]string
	decodeBody(t, resp, &result)
	return result["token"]
}

func createRoom(t *testing.T, token, name string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"name": name})
	resp := doRequest(t, "POST", baseURL+"/rooms/create", &token, body)
	assertStatus(t, resp, http.StatusCreated)

	var result map[string]interface{}
	decodeBody(t, resp, &result)
	room := result["room"].(map[string]interface{})
	return room["id"].(string)
}

func createSchedule(t *testing.T, token, roomID string, days []int, start, end string) {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{
		"daysOfWeek": days,
		"startTime":  start,
		"endTime":    end,
	})
	resp := doRequest(t, "POST", fmt.Sprintf("%s/rooms/%s/schedule/create", baseURL, roomID), &token, body)
	assertStatus(t, resp, http.StatusCreated)
}

func listSlots(t *testing.T, token, roomID, date string) []map[string]interface{} {
	t.Helper()
	resp := doRequest(t, "GET", fmt.Sprintf("%s/rooms/%s/slots/list?date=%s", baseURL, roomID, date), &token, nil)
	assertStatus(t, resp, http.StatusOK)

	var result map[string]interface{}
	decodeBody(t, resp, &result)

	slotsRaw := result["slots"].([]interface{})
	slots := make([]map[string]interface{}, len(slotsRaw))
	for i, s := range slotsRaw {
		slots[i] = s.(map[string]interface{})
	}
	return slots
}

func createBooking(t *testing.T, token, slotID string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"slotId": slotID})
	resp := doRequest(t, "POST", baseURL+"/bookings/create", &token, body)
	assertStatus(t, resp, http.StatusCreated)

	var result map[string]interface{}
	decodeBody(t, resp, &result)
	booking := result["booking"].(map[string]interface{})
	return booking["id"].(string)
}

func listMyBookings(t *testing.T, token string) []map[string]interface{} {
	t.Helper()
	resp := doRequest(t, "GET", baseURL+"/bookings/my", &token, nil)
	assertStatus(t, resp, http.StatusOK)

	var result map[string]interface{}
	decodeBody(t, resp, &result)

	raw := result["bookings"].([]interface{})
	bookings := make([]map[string]interface{}, len(raw))
	for i, b := range raw {
		bookings[i] = b.(map[string]interface{})
	}
	return bookings
}

func cancelBooking(t *testing.T, token, bookingID string) map[string]interface{} {
	t.Helper()
	resp := doRequest(t, "POST", fmt.Sprintf("%s/bookings/%s/cancel", baseURL, bookingID), &token, nil)
	assertStatus(t, resp, http.StatusOK)

	var result map[string]interface{}
	decodeBody(t, resp, &result)
	return result["booking"].(map[string]interface{})
}

func doRequest(t *testing.T, method, url string, token *string, body []byte) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != nil {
		req.Header.Set("Authorization", "Bearer "+*token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d: %s", expected, resp.StatusCode, string(body))
	}
}

func decodeBody(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
}
