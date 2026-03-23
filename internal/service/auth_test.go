package service

import (
	"testing"
)

func TestDummyLogin_Admin(t *testing.T) {
	svc := NewAuthService("test-secret")

	token, err := svc.DummyLogin("admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	userID, role, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("unexpected error parsing token: %v", err)
	}
	if role != "admin" {
		t.Errorf("expected role 'admin', got %q", role)
	}
	if userID != AdminUserID {
		t.Errorf("expected admin UUID %s, got %s", AdminUserID, userID)
	}
}

func TestDummyLogin_User(t *testing.T) {
	svc := NewAuthService("test-secret")

	token, err := svc.DummyLogin("user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	userID, role, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("unexpected error parsing token: %v", err)
	}
	if role != "user" {
		t.Errorf("expected role 'user', got %q", role)
	}
	if userID != DummyUserID {
		t.Errorf("expected user UUID %s, got %s", DummyUserID, userID)
	}
}

func TestDummyLogin_InvalidRole(t *testing.T) {
	svc := NewAuthService("test-secret")

	_, err := svc.DummyLogin("superadmin")
	if err == nil {
		t.Error("expected error for invalid role")
	}
}

func TestParseToken_Invalid(t *testing.T) {
	svc := NewAuthService("test-secret")

	_, _, err := svc.ParseToken("invalid-token-string")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	svc1 := NewAuthService("secret-1")
	svc2 := NewAuthService("secret-2")

	token, err := svc1.DummyLogin("admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, _, err = svc2.ParseToken(token)
	if err == nil {
		t.Error("expected error when parsing with wrong secret")
	}
}

func TestGenerateToken_ContainsClaims(t *testing.T) {
	svc := NewAuthService("test-secret")

	token, err := svc.GenerateToken(AdminUserID, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	userID, role, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != AdminUserID {
		t.Errorf("expected %s, got %s", AdminUserID, userID)
	}
	if role != "admin" {
		t.Errorf("expected 'admin', got %q", role)
	}
}

func TestDummyLogin_StableUUIDs(t *testing.T) {
	svc := NewAuthService("test-secret")

	token1, _ := svc.DummyLogin("user")
	token2, _ := svc.DummyLogin("user")

	uid1, _, _ := svc.ParseToken(token1)
	uid2, _, _ := svc.ParseToken(token2)

	if uid1 != uid2 {
		t.Errorf("expected stable UUID for same role, got %s and %s", uid1, uid2)
	}
}
