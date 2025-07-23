package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startTestMetricsServer(port int, statusCode int) func() {
	handler := http.NewServeMux()
	handler.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte("metrics test"))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		panic(fmt.Sprintf("failed to start test server on port %d: %v", port, err))
	}

	go server.Serve(ln)

	return func() {
		server.Close()
	}
}

func TestHealthHandler_AllOK(t *testing.T) {
	const port = 2222
	cleanup := startTestMetricsServer(port, http.StatusOK)
	defer cleanup()

	cfg := &Config{
		Web: WebConfig{
			Host: "localhost",
			Port: port,
		},
		Rabbit: RabbitConfig{
			Host:     "localhost",
			Port:     5672,
			Username: "guest",
			Password: "guest",
			Channel:  "",
		},
	}

	time.Sleep(100 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	HealthHandler(rec, req, cfg)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestHealthHandler_MetricsFails(t *testing.T) {
	const port = 2222
	cleanup := startTestMetricsServer(port, http.StatusInternalServerError)
	defer cleanup()

	cfg := &Config{
		Web: WebConfig{
			Host: "localhost",
			Port: port,
		},
		Rabbit: RabbitConfig{
			Host:     "localhost",
			Port:     5672,
			Username: "guest",
			Password: "guest",
			Channel:  "",
		},
	}

	time.Sleep(100 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	HealthHandler(rec, req, cfg)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503 Service Unavailable for metrics failure, got %d", resp.StatusCode)
	}
}

func TestHealthHandler_RabbitFails(t *testing.T) {
	const port = 2222
	cleanup := startTestMetricsServer(port, http.StatusOK)
	defer cleanup()

	cfg := &Config{
		Web: WebConfig{
			Host: "localhost",
			Port: port,
		},
		Rabbit: RabbitConfig{
			Host:     "localhost",
			Port:     5999, 
			Username: "guest",
			Password: "guest",
			Channel:  "",
		},
	}

	time.Sleep(100 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	HealthHandler(rec, req, cfg)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503 Service Unavailable for RabbitMQ failure, got %d", resp.StatusCode)
	}
}

func TestHealthHandler_MetricsConnectionFails(t *testing.T) {
	cfg := &Config{
		Web: WebConfig{
			Host: "localhost",
			Port: 5999, 
		},
		Rabbit: RabbitConfig{
			Host:     "localhost",
			Port:     5672,
			Username: "guest",
			Password: "guest",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	HealthHandler(rec, req, cfg)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503 Service Unavailable for metrics connection failure, got %d", resp.StatusCode)
	}
}
