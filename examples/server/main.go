// Command server queues the body of each POST /jobs request for four background workers.
// It answers 202 Accepted once the job is queued, and 503 Service Unavailable when the queue stays full
// for 100 ms, so that clients back off. On Ctrl-C it stops taking requests and finishes the queued jobs.
//
//	go run ./examples/server
//	curl -d 'resize photo.jpg' localhost:8080/jobs
package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/vardius/worker-pool/v3"
)

// enqueue queues each request's body as a job, waiting up to wait for room in the queue.
func enqueue(pool *workerpool.Pool[string], wait time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), wait)
		defer cancel()
		if err := pool.Delegate(ctx, string(body)); err != nil {
			// The queue stayed full, the client went away, or the server is shutting down.
			w.Header().Set("Retry-After", "1")
			http.Error(w, "busy, try again later", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

func process(job string) {
	time.Sleep(time.Second) // stands in for real work
	log.Printf("done: %s", job)
}

func main() {
	pool := workerpool.New[string](100)
	for range 4 {
		pool.AddWorker(process)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", enqueue(pool, 100*time.Millisecond))
	srv := &http.Server{Addr: "localhost:8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.ListenAndServe() }()
	select {
	case err := <-serveErr:
		log.Fatal(err)
	case <-ctx.Done():
	}

	// Stop taking requests first, so that no handler is left queueing when the pool stops.
	log.Print("shutting down, finishing the queued jobs")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Print(err)
	}
	pool.Stop()
}
