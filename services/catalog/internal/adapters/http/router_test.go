package httpadapter

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	ok := ReadinessCheck{Name: "ok", Check: func(context.Context) error { return nil }}
	broken := ReadinessCheck{Name: "postgres", Check: func(context.Context) error {
		return errors.New("connection refused")
	}}

	tests := []struct {
		name     string
		path     string
		checks   []ReadinessCheck
		wantCode int
		wantBody string
	}{
		{"liveness ignores dependencies", "/healthz", []ReadinessCheck{broken}, http.StatusOK, `{"status":"ok"}`},
		{"ready without checks", "/readyz", nil, http.StatusOK, `{"status":"ready"}`},
		{"ready when all checks pass", "/readyz", []ReadinessCheck{ok}, http.StatusOK, `{"status":"ready"}`},
		{"not ready when a check fails", "/readyz", []ReadinessCheck{ok, broken},
			http.StatusServiceUnavailable, `{"failed":["postgres"],"status":"not_ready"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter(logger, tt.checks...)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantCode)
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Fatalf("body = %s, want %s", got, tt.wantBody)
			}
		})
	}
}
