package credentials

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestProtonPassResolver_ProviderName(t *testing.T) {
	r := NewProtonPassResolver()
	if name := r.ProviderName(); name != "Proton Pass" {
		t.Errorf("expected ProviderName 'Proton Pass', got %q", name)
	}
}

func TestProtonPassResolver_CanResolve(t *testing.T) {
	r := NewProtonPassResolver()

	tests := []struct {
		uri      string
		expected bool
	}{
		{"pass://Vault/Item/Field", true},
		{" pass://Vault/Item ", true},
		{"env://MY_VAR", false},
		{"https://proton.me", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := r.CanResolve(tt.uri); got != tt.expected {
			t.Errorf("CanResolve(%q) = %v; want %v", tt.uri, got, tt.expected)
		}
	}
}

func TestProtonPassResolver_IsAvailable(t *testing.T) {
	r := NewProtonPassResolver()

	// Case 1: Binary found
	r.LookPath = func(file string) (string, error) {
		if file == "pass-cli" {
			return "/usr/local/bin/pass-cli", nil
		}
		return "", errors.New("not found")
	}
	if !r.IsAvailable() {
		t.Errorf("expected IsAvailable() to be true when pass-cli is in path")
	}

	// Case 2: Binary missing
	r.LookPath = func(file string) (string, error) {
		return "", errors.New("not found")
	}
	if r.IsAvailable() {
		t.Errorf("expected IsAvailable() to be false when pass-cli is missing")
	}
}

func TestProtonPassResolver_Resolve(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid URI scheme", func(t *testing.T) {
		r := NewProtonPassResolver()
		_, err := r.Resolve(ctx, "invalid://uri")
		if !errors.Is(err, ErrInvalidURI) {
			t.Fatalf("expected ErrInvalidURI, got %v", err)
		}
	})

	t.Run("pass-cli binary missing", func(t *testing.T) {
		r := NewProtonPassResolver()
		r.LookPath = func(file string) (string, error) {
			return "", errors.New("not found")
		}
		_, err := r.Resolve(ctx, "pass://Vault/Item/Field")
		if !errors.Is(err, ErrBinaryNotFound) {
			t.Fatalf("expected ErrBinaryNotFound, got %v", err)
		}
	})

	t.Run("successful secret resolution", func(t *testing.T) {
		r := NewProtonPassResolver()
		r.LookPath = func(file string) (string, error) {
			return "/usr/bin/pass-cli", nil
		}
		r.Exec = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			if name == "pass-cli" && len(args) == 3 && args[0] == "item" && args[1] == "view" && args[2] == "pass://Vault/Item/Field" {
				return []byte("secret_api_key_12345\n"), nil
			}
			return nil, errors.New("unexpected command execution")
		}

		secret, err := r.Resolve(ctx, "pass://Vault/Item/Field")
		if err != nil {
			t.Fatalf("unexpected error resolving secret: %v", err)
		}
		if secret != "secret_api_key_12345" {
			t.Errorf("expected resolved secret 'secret_api_key_12345', got %q", secret)
		}
	})

	t.Run("command execution error / item missing or unauthenticated", func(t *testing.T) {
		r := NewProtonPassResolver()
		r.LookPath = func(file string) (string, error) {
			return "/usr/bin/pass-cli", nil
		}
		r.Exec = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return nil, errors.New("item not found")
		}

		_, err := r.Resolve(ctx, "pass://Vault/Item/MissingField")
		if err == nil || !errors.Is(err, ErrResolutionFailed) {
			t.Fatalf("expected ErrResolutionFailed, got %v", err)
		}
	})

	t.Run("context timeout handling", func(t *testing.T) {
		r := NewProtonPassResolver()
		r.Timeout = 10 * time.Millisecond
		r.LookPath = func(file string) (string, error) {
			return "/usr/bin/pass-cli", nil
		}
		r.Exec = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(50 * time.Millisecond):
				return []byte("late response"), nil
			}
		}

		_, err := r.Resolve(ctx, "pass://Vault/Item/SlowField")
		if err == nil {
			t.Fatalf("expected context timeout error, got nil")
		}
	})
}
