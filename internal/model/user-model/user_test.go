package user_model

import (
	"testing"

	"github.com/google/uuid"
)

func TestMustParseUUID_Valid(t *testing.T) {
	id := "550e8400-e29b-41d4-a716-446655440000"
	parsed := MustParseUUID(id)
	if parsed.String() != id {
		t.Fatalf("MustParseUUID(%q) = %q, want %q", id, parsed.String(), id)
	}
}

func TestMustParseUUID_PanicsOnInvalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustParseUUID should panic on invalid UUID")
		}
	}()
	MustParseUUID("not-a-uuid")
}

func TestMustParseUUID_ReturnsUUIDType(t *testing.T) {
	id := uuid.New()
	result := MustParseUUID(id.String())
	if result != id {
		t.Fatalf("MustParseUUID mismatch: got %v, want %v", result, id)
	}
}
