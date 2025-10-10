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
	if resp.Header.Get("X-Cache") != "" {
		t.Errorf("expected no header X-Cache=MISS, got %s", resp.Header.Get("X-Cache"))
	}
}

func TestHandleRequestCacheHit(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("X-Backend", "ok")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "from backend: "+string(body))
	}))
	defer backend.Close()

	cfg := config.NewConfig(":0", map[string]config.Route{
		"/api": {
			LoadBalancerType: config.LBStrategySingle,
			CacheConfig: config.CacheConfig{
				Enabled:      true,
				TTL:          60,
				MaxSize:      1, // MiB
				MaxEntrySize: 1, // MiB
			},
			Backends: []config.Backend{
				{URL: backend.URL},
			},
		},
	})

	rev := NewReverser(cfg)
	ts := httptest.NewServer(http.HandlerFunc(rev.handleRequest))
	defer ts.Close()

	req1 := httptest.NewRequest("GET", ts.URL+"/api", nil)
	w1 := httptest.NewRecorder()
	rev.handleRequest(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}
	if w1.Header().Get("X-Cache") != "MISS" {
		t.Errorf("expected X-Cache=MISS, got %s", w1.Header().Get("X-Cache"))
	}

	body1 := w1.Body.String()
	if !strings.Contains(body1, "from backend") {
		t.Errorf("unexpected body: %s", body1)
	}

	req2 := httptest.NewRequest("GET", ts.URL+"/api", nil)
	w2 := httptest.NewRecorder()
	rev.handleRequest(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	if w2.Header().Get("X-Cache") != "HIT" {
		t.Errorf("expected X-Cache=HIT, got %s", w2.Header().Get("X-Cache"))
	}

	body2 := w2.Body.String()
	if body1 != body2 {
		t.Errorf("expected cached body to match, got:\n%s\nvs\n%s", body1, body2)
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
