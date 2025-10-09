package reverser

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/papey/cmiyc/internal/config"
	"github.com/papey/cmiyc/internal/forwarder"
)

func makeConfig(listen, backendURL string) config.Config {
	return config.NewConfig(listen, map[string]config.Route{
		"/api": {
			LoadBalancerType: config.LBStrategySingle,
			Backends: []config.Backend{
				{URL: backendURL},
			},
		},
	})
}

func TestHandleRequestSuccess(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("X-Backend", "ok")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "from backend: "+string(body))
	}))
	defer backend.Close()

	cfg := makeConfig(":0", backend.URL)
	rev := NewReverser(cfg)

	ts := httptest.NewServer(http.HandlerFunc(rev.handleRequest))
	defer ts.Close()

	reqBody := strings.NewReader("hello reverser")
	resp, err := http.Post(ts.URL+"/api", "text/plain", reqBody)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	data, _ := io.ReadAll(resp.Body)
	body := string(data)

	if !strings.Contains(body, "from backend") {
		t.Errorf("unexpected body: %s", body)
	}
	if resp.Header.Get("X-Backend") != "ok" {
		t.Errorf("expected header X-Backend=ok, got %s", resp.Header.Get("X-Backend"))
	}
}

func TestHandleRequestRouteNotFound(t *testing.T) {
	cfg := makeConfig(":0", "http://localhost")
	rev := NewReverser(cfg)
	rev.client = forwarder.NewClient()

	req := httptest.NewRequest("GET", "/unknown", nil)
	w := httptest.NewRecorder()

	rev.handleRequest(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestStartAndStop(t *testing.T) {
	cfg := makeConfig(":0", "http://localhost")
	rev := NewReverser(cfg)
	rev.client = forwarder.NewClient()

	go func() {
		_ = rev.Start()
	}()
	time.Sleep(50 * time.Millisecond)

	if rev.server == nil {
		t.Fatal("expected server to be initialized")
	}

	_, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := rev.Stop(); err != nil {
		t.Fatalf("expected no error on Stop, got %v", err)
	}
}
