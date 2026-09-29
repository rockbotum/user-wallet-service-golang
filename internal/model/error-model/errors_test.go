package error_model

import (
	"errors"
	"testing"
)

func TestErrorWrapping(t *testing.T) {
	cause := errors.New("db: connection refused")
	err := Wrap(KindInternal, "failed to query", cause)

	if !errors.Is(err, cause) {
		t.Fatal("errors.Is should unwrap to the cause")
	}
	if !Is(err, KindInternal) {
		t.Fatalf("Is(KindInternal) = false, want true")
	}
	if Is(err, KindNotFound) {
		t.Fatalf("Is(KindNotFound) = true, want false")
	}
}

func TestErrorMessages(t *testing.T) {
	cause := errors.New("root cause")

	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{name: "with message", err: New(KindNotFound, "user not found"), want: "user not found"},
		{name: "wrapped with message", err: Wrap(KindInvalid, "bad input", cause), want: "bad input"},
		{name: "wrapped without message", err: &Error{Kind: KindInternal, Err: cause}, want: "root cause"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Fatalf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
