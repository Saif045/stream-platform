package user

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"stream-platform/internal/auth"
)

type fakeUserStore struct {
	users map[string]*User

	createErr error
	getErr    error

	created *User

	getByUsernameCalledWith string
}

func (f *fakeUserStore) Create(ctx context.Context, user *User) error {
	if f.createErr != nil {
		return f.createErr
	}

	if f.users == nil {
		f.users = make(map[string]*User)
	}

	f.created = user
	f.users[user.Username] = user

	return nil
}

func (f *fakeUserStore) GetByUsername(ctx context.Context, username string) (*User, error) {
	f.getByUsernameCalledWith = username

	if f.getErr != nil {
		return nil, f.getErr
	}

	user, ok := f.users[username]
	if !ok {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func TestRegister(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		store := &fakeUserStore{}
		service := NewService(store)

		user, err := service.Register(context.Background(), " seif ", "password123")
		if err != nil {
			t.Fatal(err)
		}

		if user.ID == "" {
			t.Fatal("expected generated user id")
		}

		if user.Username != "seif" {
			t.Fatalf("expected username %q, got %q", "seif", user.Username)
		}

		if user.PasswordHash == "" {
			t.Fatal("expected password hash")
		}

		if user.PasswordHash == "password123" {
			t.Fatal("expected password hash, got raw password")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")); err != nil {
			t.Fatal("expected password hash to match original password")
		}

		if store.created != user {
			t.Fatal("expected user to be passed to store")
		}
	})

	t.Run("rejects empty username", func(t *testing.T) {
		store := &fakeUserStore{}
		service := NewService(store)

		_, err := service.Register(context.Background(), "   ", "password123")
		if err == nil {
			t.Fatal("expected error")
		}

		if store.created != nil {
			t.Fatal("expected store not to be called")
		}
	})

	t.Run("rejects short password", func(t *testing.T) {
		store := &fakeUserStore{}
		service := NewService(store)

		_, err := service.Register(context.Background(), "seif", "short")
		if err == nil {
			t.Fatal("expected error")
		}

		if store.created != nil {
			t.Fatal("expected store not to be called")
		}
	})

	t.Run("returns store error", func(t *testing.T) {
		storeErr := errors.New("store failed")

		store := &fakeUserStore{createErr: storeErr}
		service := NewService(store)

		_, err := service.Register(context.Background(), "seif", "password123")
		if !errors.Is(err, storeErr) {
			t.Fatalf("expected store error, got %v", err)
		}
	})
}

func TestLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		auth.SetSecret("test-secret")

		passwordHash := mustHashPassword(t, "password123")

		store := &fakeUserStore{
			users: map[string]*User{
				"seif": {
					ID:           "user-1",
					Username:     "seif",
					PasswordHash: passwordHash,
				},
			},
		}

		service := NewService(store)

		token, err := service.Login(context.Background(), " seif ", "password123")
		if err != nil {
			t.Fatal(err)
		}

		if token == "" {
			t.Fatal("expected token")
		}

		if store.getByUsernameCalledWith != "seif" {
			t.Fatalf("expected username lookup %q, got %q", "seif", store.getByUsernameCalledWith)
		}
	})

	t.Run("rejects missing user", func(t *testing.T) {
		store := &fakeUserStore{}
		service := NewService(store)

		_, err := service.Login(context.Background(), "missing-user", "password123")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("rejects wrong password", func(t *testing.T) {
		passwordHash := mustHashPassword(t, "password123")

		store := &fakeUserStore{
			users: map[string]*User{
				"seif": {
					ID:           "user-1",
					Username:     "seif",
					PasswordHash: passwordHash,
				},
			},
		}

		service := NewService(store)

		_, err := service.Login(context.Background(), "seif", "wrong-password")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("rejects invalid password hash", func(t *testing.T) {
		store := &fakeUserStore{
			users: map[string]*User{
				"seif": {
					ID:           "user-1",
					Username:     "seif",
					PasswordHash: "not-a-valid-bcrypt-hash",
				},
			},
		}

		service := NewService(store)

		_, err := service.Login(context.Background(), "seif", "password123")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("hides store error as invalid credentials", func(t *testing.T) {
		store := &fakeUserStore{getErr: errors.New("database failed")}
		service := NewService(store)

		_, err := service.Login(context.Background(), "seif", "password123")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	return string(hash)
}
