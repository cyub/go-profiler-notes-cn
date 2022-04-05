package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	if err := os.Setenv("DD_PROFILING_WAIT_PROFILE", "yes"); err != nil {
		return err
	} else if err := profiler.Start(
		profiler.WithService("stuck-deadlock"),
		profiler.WithVersion(time.Now().String()),
		profiler.WithPeriod(60*time.Second),
		profiler.WithProfileTypes(
			profiler.GoroutineProfile,
		),
	); err != nil {
		return err
	}

	var (
		a = &sync.Mutex{}
		b = &sync.Mutex{}
	)
	go bob(a, b)
	go alice(a, b)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// force gc on a regular basis to make sure g.waitsince gets populated.
			fmt.Printf("gc\n")
			runtime.GC()
		case sig := <-c:
			fmt.Printf("sig: %s\n", sig)
			return nil
		}
	}
	return nil
}

func bob(a, b *sync.Mutex) {
	for {
		fmt.Println("bob is okay")
		a.Lock()
		b.Lock()
		// do stuff
		a.Unlock()
		b.Unlock()
	}
}

func alice(a, b *sync.Mutex) {
	for {
		fmt.Println("alice is okay")
		b.Lock()
		a.Lock()
		// do stuff
		b.Unlock()
		a.Unlock()
	}
}
