package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware(t *testing.T) {
	t.Run("rejects missing authorization header", func(t *testing.T) {
		called := false

		handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

		if called {
			t.Fatal("expected next handler not to be called")
		}
	})

	t.Run("rejects malformed authorization header", func(t *testing.T) {
		called := false

		handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Token abc123")

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

		if called {
			t.Fatal("expected next handler not to be called")
		}
	})

	t.Run("rejects empty bearer token", func(t *testing.T) {
		called := false

		handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer ")

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

		if called {
			t.Fatal("expected next handler not to be called")
		}
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		called := false

		handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

		if called {
			t.Fatal("expected next handler not to be called")
		}
	})

	t.Run("valid token reaches next handler with user id in context", func(t *testing.T) {
		SetSecret("test-secret")

		token, err := GenerateToken("user-1")
		if err != nil {
			t.Fatal(err)
		}

		called := false
		var gotUserID string

		handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true

			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				t.Fatal("expected user id in context")
			}

			gotUserID = userID

			w.WriteHeader(http.StatusNoContent)
		}))

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}

		if !called {
			t.Fatal("expected next handler to be called")
		}

		if gotUserID != "user-1" {
			t.Fatalf("expected user id %q, got %q", "user-1", gotUserID)
		}
	})
}
