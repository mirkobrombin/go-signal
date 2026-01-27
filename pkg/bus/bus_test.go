package bus_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mirkobrombin/go-signal/v2/pkg/bus"
)

type Event struct {
	Greeting string
}

type IEvent interface {
	GetGreeting() string
}

func (e *Event) GetGreeting() string {
	return e.Greeting
}

func TestBus_DirectType(t *testing.T) {
	b := bus.New()
	received := false

	bus.Subscribe(b, func(ctx context.Context, e *Event) error {
		if e.Greeting != "Hello" {
			return errors.New("wrong greeting")
		}
		received = true
		return nil
	})

	err := bus.Emit(context.Background(), b, &Event{Greeting: "Hello"})
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	if !received {
		t.Fatal("Handler not called")
	}
}

func TestBus_Interface(t *testing.T) {
	b := bus.New()
	received := false

	// Subscribe to Interface
	bus.Subscribe(b, func(ctx context.Context, e IEvent) error {
		if e.GetGreeting() != "Hello" {
			return errors.New("wrong greeting")
		}
		received = true
		return nil
	})

	// Emit Interface
	var evt IEvent = &Event{Greeting: "Hello"}
	err := bus.Emit(context.Background(), b, evt)
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	if !received {
		t.Fatal("Handler not called")
	}
}

func TestBus_MismatchedTypes(t *testing.T) {
	b := bus.New()
	received := false

	// Subscribe to Concrete
	bus.Subscribe(b, func(ctx context.Context, e *Event) error {
		received = true
		return nil
	})

	// Emit Interface (holding Concrete) - Should NOT trigger Concrete handler
	// Because TypeFor[IEvent] != TypeFor[*Event]
	var evt IEvent = &Event{Greeting: "Hello"}
	_ = bus.Emit(context.Background(), b, evt)

	if received {
		t.Fatal("Handler called incorrectly (interface emit triggered concrete handler)")
	}

	// But Emit Concrete should work
	_ = bus.Emit(context.Background(), b, &Event{Greeting: "Hello"})
	if !received {
		t.Fatal("Handler not called for concrete emit")
	}
}

func TestBus_StopOnFirstError(t *testing.T) {
	b := bus.New(bus.WithStrategy(bus.StopOnFirstError))

	bus.Subscribe(b, func(ctx context.Context, e *Event) error {
		return errors.New("error 1")
	}, bus.PriorityHigh)

	bus.Subscribe(b, func(ctx context.Context, e *Event) error {
		t.Fatal("Should not be called")
		return nil
	}, bus.PriorityLow)

	err := bus.Emit(context.Background(), b, &Event{})
	if err == nil || err.Error() != "error 1" {
		t.Fatalf("Expected 'error 1', got %v", err)
	}
}

func TestBus_BestEffort(t *testing.T) {
	b := bus.New(bus.WithStrategy(bus.BestEffort))
	calls := 0

	bus.Subscribe(b, func(ctx context.Context, e *Event) error {
		calls++
		return errors.New("error 1")
	}, bus.PriorityHigh)

	bus.Subscribe(b, func(ctx context.Context, e *Event) error {
		calls++
		return errors.New("error 2")
	}, bus.PriorityLow)

	err := bus.Emit(context.Background(), b, &Event{})
	if err == nil {
		t.Fatal("Expected error")
	}

	// errors.Join might format differently
	// Check call count
	if calls != 2 {
		t.Fatalf("Expected 2 calls, got %d", calls)
	}
}

func TestBus_Async(t *testing.T) {
	b := bus.New()
	wg := sync.WaitGroup{}
	wg.Add(1)

	bus.Subscribe(b, func(ctx context.Context, e *Event) error {
		defer wg.Done()
		return nil
	})

	bus.EmitAsync(context.Background(), b, &Event{})

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for async")
	}
}
