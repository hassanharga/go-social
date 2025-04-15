package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github/hassanharga/go-social/internal/auth"
	// "github/hassanharga/go-social/internal/ratelimiter"
	"github/hassanharga/go-social/internal/store"
	"github/hassanharga/go-social/internal/store/cache"
)

func newTestApplication(t *testing.T, cfg config) *application {
	t.Helper()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	// logger := zap.NewNop().Sugar()
	// Uncomment to enable logs
	// logger := zap.Must(zap.NewProduction()).Sugar()
	mockStore := store.NewMockStore()
	mockCacheStore := cache.NewMockStore()

	testAuth := &auth.TestAuthenticator{}

	// Rate limiter
	// rateLimiter := ratelimiter.NewFixedWindowLimiter(
	// 	cfg.rateLimiter.RequestsPerTimeFrame,
	// 	cfg.rateLimiter.TimeFrame,
	// )

	return &application{
		logger:        logger,
		store:         mockStore,
		cacheStorage:  mockCacheStore,
		authenticator: testAuth,
		config:        cfg,
		// rateLimiter:   rateLimiter,
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d", expected, actual)
	}
}
