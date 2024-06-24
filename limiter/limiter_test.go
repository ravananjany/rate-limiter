package limiter

import (
	"context"
	"limiter/lib"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockRateLimiter struct {
	IsAllow bool
}

func (m *mockRateLimiter) Allow() bool {
	return m.IsAllow
}
func (m *mockRateLimiter) Wait(context.Context) error {
	return nil
}

func TestLimit(t *testing.T) {
	tests := []struct {
		name      string
		clientsIp string
		rps       float64
		burst     int
		args      lib.RateLimiterInterface
		params    func(writer http.ResponseWriter, request *http.Request)
		failure   bool
	}{
		{name: "success request", args: &mockRateLimiter{true}, params: func(writer http.ResponseWriter, request *http.Request) {}, clientsIp: "127.0.0.1"},
		{name: "drop request", args: &mockRateLimiter{false}, params: func(writer http.ResponseWriter, request *http.Request) {}, clientsIp: "127.0.0.2"},
		{name: "ip not found request", args: &mockRateLimiter{false}, params: func(writer http.ResponseWriter, request *http.Request) {}, clientsIp: ""},
		{name: "ip error", args: &mockRateLimiter{false}, params: func(writer http.ResponseWriter, request *http.Request) {}, clientsIp: "192.2.", failure: true},
	}

	for _, tt := range tests {
		clients[tt.clientsIp] = &dynamicConfigurableLimiter{rps: tt.rps, burst: tt.burst, limiterlib: &mockRateLimiter{tt.args.Allow()}}
		t.Run(tt.name, func(t *testing.T) {
			handler := Limit(tt.params)
			req := httptest.NewRequest("GET", "http://google.com", nil)
			w := httptest.NewRecorder()
			if tt.clientsIp != "" {
				req.RemoteAddr = tt.clientsIp + ":1234"
			}
			if tt.failure {
				req.RemoteAddr = tt.clientsIp
			}
			handler.ServeHTTP(w, req)
		})
		clients = map[string]DynamicConfigurableLimiter{}
	}

}

func TestNewNewConfigurableLimiter(t *testing.T) {
	dynamicLimiter := NewConfigurableLimiter(2, 4)
	assert.NotNil(t, dynamicLimiter)
}

func TestAllow(t *testing.T) {
	dynamicLimiter := dynamicConfigurableLimiter{limiterlib: &mockRateLimiter{}}
	dynamicLimiter.Allow()
}

func TestWait(t *testing.T) {
	dynamicLimiter := dynamicConfigurableLimiter{limiterlib: &mockRateLimiter{}}
	dynamicLimiter.Wait(context.Background())
}

func TestLoadConfig(t *testing.T) {
	LoadClients()
}
