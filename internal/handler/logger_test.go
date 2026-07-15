package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danilov-go/gophermart/internal/handler"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestLogger(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		method      string
		handlerFunc func(w http.ResponseWriter, r *http.Request)
		status      int64
		size        int64
	}{
		{
			name:   "GET запрос",
			url:    "/api/user/balance",
			method: http.MethodGet,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`test`))
			},
			status: http.StatusOK,
			size:   4,
		},
		{
			name:   "POST запрос",
			url:    "/api/user/withdraw",
			method: http.MethodPost,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusAccepted)
			},
			status: http.StatusAccepted,
			size:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zapcore.InfoLevel)
			testLogger := zap.New(core)
			middleware := handler.RequestLogger(testLogger)
			h := middleware(http.HandlerFunc(tt.handlerFunc))
			req := httptest.NewRequest(tt.method, tt.url, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			logEntry := logs.All()[0]
			message := logEntry.ContextMap()
			assert.Equal(t, tt.url, message["url"])
			assert.Equal(t, tt.method, message["method"])
			assert.Equal(t, tt.status, message["status"])
			assert.Equal(t, tt.size, message["size"])
			assert.Contains(t, message, "duration")
		})
	}
}
