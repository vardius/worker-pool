// Command checksum prints the SHA-256 of every file named on its command line, hashing one file per CPU at once.
//
//	go run ./examples/checksum go.mod README.md
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"

	"github.com/vardius/worker-pool/v3"
)

// A sum is a file's SHA-256, or the error that kept it from being read.
type sum struct {
	hex string
	err error
}

// checksums hashes the files on a pool of workers; sums[i] is for paths[i]. Once ctx is done,
// it queues no more files, waits for the ones already queued, and returns ctx's error.
func checksums(ctx context.Context, paths []string, workers int) ([]sum, error) {
	sums := make([]sum, len(paths))

	// A job is the index of a path. Each one is run once, by one worker, so the writes to sums need no lock.
	pool := workerpool.New[int](workers)
	for range workers {
		pool.AddWorker(func(i int) { sums[i].hex, sums[i].err = hashFile(paths[i]) })
	}
	for i := range paths {
		if err := pool.Delegate(ctx, i); err != nil {
			pool.Stop()
			return nil, err
		}
	}
	pool.Stop() // returns once every queued file is hashed

	return sums, nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func main() {
	log.SetFlags(0)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	context.AfterFunc(ctx, stop) // after the first Ctrl-C, a second one ends the program at once

	paths := os.Args[1:]
	sums, err := checksums(ctx, paths, runtime.NumCPU())
	if err != nil {
		log.Fatal(err)
	}

	failed := false
	for i, s := range sums {
		if s.err != nil {
			log.Print(s.err)
			failed = true
			continue
		}
		fmt.Printf("%s  %s\n", s.hex, paths[i])
	}
	if failed {
		os.Exit(1)
	}
}
