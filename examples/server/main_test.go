package main

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/vardius/worker-pool/v3"
)

func TestEnqueue(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		pool := workerpool.New[string](1)
		release := make(chan struct{})
		var mu sync.Mutex
		var done []string
		pool.AddWorker(func(job string) {
			<-release
			mu.Lock()
			defer mu.Unlock()
			done = append(done, job)
		})
		handler := enqueue(pool, 100*time.Millisecond)

		steps := []struct {
			name, body string
			before     func()
			wantCode   int
			wantWait   time.Duration
		}{
			{"the worker takes it", "a", func() {}, http.StatusAccepted, 0},
			{"it waits in the queue", "b", synctest.Wait, http.StatusAccepted, 0},
			{"the queue stays full", "c", func() {}, http.StatusServiceUnavailable, 100 * time.Millisecond},
			{"the pool is stopped", "d", func() { close(release); pool.Stop() }, http.StatusServiceUnavailable, 0},
		}
		for _, step := range steps {
			step.before()
			start := time.Now()
			rec := httptest.NewRecorder()
			handler(rec, httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(step.body)))

			if rec.Code != step.wantCode {
				t.Errorf("%s: status = %d, want %d", step.name, rec.Code, step.wantCode)
			}
			if waited := time.Since(start); waited != step.wantWait {
				t.Errorf("%s: answered after %v, want %v", step.name, waited, step.wantWait)
			}
			if rec.Code == http.StatusServiceUnavailable && rec.Header().Get("Retry-After") == "" {
				t.Errorf("%s: no Retry-After header", step.name)
			}
		}

		if want := []string{"a", "b"}; !slices.Equal(done, want) {
			t.Errorf("ran %v, want %v", done, want)
		}
	})
}
