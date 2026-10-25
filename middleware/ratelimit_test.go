package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ThembinkosiThemba/zen"
)

func TestSlidingWindow(t *testing.T) {
	rl := newRateLimiter(4, time.Minute)
	start := time.Now()

	for i := range 4 {
		if ok, _ := rl.allow("k", start); !ok {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	ok, retry := rl.allow("k", start)
	if ok || retry <= 0 || retry > time.Minute {
		t.Fatalf("5th request: ok=%v retry=%v", ok, retry)
	}
	if ok, _ := rl.allow("other", start); !ok {
		t.Fatal("keys must be limited independently")
	}

	// Halfway into the next window, half of the previous 4 still count.
	mid := start.Add(time.Minute + 30*time.Second)
	for i := range 2 {
		if ok, _ := rl.allow("k", mid); !ok {
			t.Fatalf("request %d in next window should be allowed", i+1)
		}
	}
	if ok, _ := rl.allow("k", mid); ok {
		t.Fatal("previous window should still count")
	}

	// After two idle windows, the full limit is available again.
	later := start.Add(5 * time.Minute)
	for i := range 4 {
		if ok, _ := rl.allow("k", later); !ok {
			t.Fatalf("request %d after idle should be allowed", i+1)
		}
	}
}

func TestRateLimiterSweepsIdleKeys(t *testing.T) {
	rl := newRateLimiter(1, time.Second)
	now := time.Now()
	rl.allow("a", now)
	rl.allow("b", now)
	rl.allow("c", now.Add(3*time.Second))
	if len(rl.entries) != 1 {
		t.Errorf("idle keys not swept: %d entries", len(rl.entries))
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	app := zen.New()
	app.Apply(RateLimitWithConfig(RateLimitConfig{Limit: 2, Window: time.Minute}))
	app.GET("/", func(c *zen.Context) { c.Text(200, "ok") })

	// Requests from the same IP on different source ports share a limit.
	var codes []int
	var w *httptest.ResponseRecorder
	for _, port := range []string{"1000", "2000", "3000"} {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.1:" + port
		req.Header.Set("X-Forwarded-For", "spoofed-"+port)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, req)
		codes = append(codes, w.Code)
	}
	if codes[0] != 200 || codes[1] != 200 || codes[2] != http.StatusTooManyRequests {
		t.Fatalf("codes = %v", codes)
	}
	if w.Header().Get("Retry-After") == "" || w.Body.String() != "rate limit exceeded" {
		t.Errorf("rejected response: headers=%v body=%q", w.Header(), w.Body.String())
	}
}

func TestRateLimitConcurrent(t *testing.T) {
	app := zen.New()
	app.Apply(RateLimitWithConfig(RateLimitConfig{Limit: 50, Window: time.Minute}))
	app.GET("/", func(c *zen.Context) {})

	var mu sync.Mutex
	allowed := 0
	var wg sync.WaitGroup
	for range 200 {
		wg.Go(func() {
			w := httptest.NewRecorder()
			app.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
			if w.Code == 200 {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if allowed != 50 {
		t.Errorf("allowed %d requests, want exactly 50", allowed)
	}
}
