package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mirkobrombin/go-signal/v2/pkg/bus"
)

type LogEvent struct {
	Message string
}

func main() {
	b := bus.New()

	// High priority logger
	bus.Subscribe(b, func(ctx context.Context, e LogEvent) error {
		fmt.Printf("[Sync] Log received: %s\n", e.Message)
		return nil
	}, bus.PriorityHigh)

	// Async listener
	bus.Subscribe(b, func(ctx context.Context, e LogEvent) error {
		time.Sleep(100 * time.Millisecond) // Simulate slow work
		fmt.Printf("[Async] Background worker finished: %s\n", e.Message)
		return nil
	})

	fmt.Println("--- Synchronous Emit ---")
	_ = bus.Emit(context.Background(), b, LogEvent{Message: "System booted"})

	fmt.Println("\n--- Asynchronous Emit ---")
	bus.EmitAsync(context.Background(), b, LogEvent{Message: "Heavy background task"})

	fmt.Println("Main thread continues immediately...")

	// Wait for async worker to finish in this example
	time.Sleep(200 * time.Millisecond)
	fmt.Println("Done.")
}
