package bus

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"sync"
)

// Handler defines the function signature for event listeners.
type Handler[T any] func(ctx context.Context, event T) error

// Priority defines the order of execution for listeners.
type Priority int

const (
	PriorityHigh   Priority = 100
	PriorityNormal Priority = 0
	PriorityLow    Priority = -100
)

// DispatchStrategy defines how the bus handles multiple listeners.
type DispatchStrategy int

const (
	StopOnFirstError DispatchStrategy = iota
	BestEffort
)

// Bus is a type-safe event bus.
type Bus struct {
	subscribers map[reflect.Type][]subscriber
	mu          sync.RWMutex
	strategy    DispatchStrategy
}

type subscriber struct {
	handler  any // Wrapped Handler[T]
	priority Priority
}

var defaultBus = New()

// Default returns the package-level default bus.
func Default() *Bus {
	return defaultBus
}

// New creates a new Bus instance.
func New(opts ...Option) *Bus {
	b := &Bus{
		subscribers: make(map[reflect.Type][]subscriber),
		strategy:    StopOnFirstError,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Option defines a functional configuration for the Bus.
type Option func(*Bus)

// WithStrategy sets the dispatch strategy for the bus.
func WithStrategy(s DispatchStrategy) Option {
	return func(b *Bus) { b.strategy = s }
}

// Subscribe registers a listener for a specific event type.
// If b is nil, it uses the default global bus.
func Subscribe[T any](b *Bus, fn Handler[T], priority ...Priority) {
	if b == nil {
		b = defaultBus
	}

	p := PriorityNormal
	if len(priority) > 0 {
		p = priority[0]
	}

	key := reflect.TypeFor[T]()

	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[key] = append(b.subscribers[key], subscriber{
		handler:  fn,
		priority: p,
	})

	sort.SliceStable(b.subscribers[key], func(i, j int) bool {
		return b.subscribers[key][i].priority > b.subscribers[key][j].priority
	})
}

// Emit dispatches an event to all registered listeners synchronously.
// If b is nil, it uses the default global bus.
func Emit[T any](ctx context.Context, b *Bus, event T) error {
	if b == nil {
		b = defaultBus
	}

	// We use reflect.TypeOf(event) to get the concrete type,
	// which works even if T is an interface{} (any).
	key := reflect.TypeOf(event)

	b.mu.RLock()
	subs, ok := b.subscribers[key]
	b.mu.RUnlock()

	if !ok {
		return nil
	}

	var errs []error
	for _, sub := range subs {
		// We use reflection to call the handler because the stored handler
		// might have a specific type (e.g. Handler[*MyEvent]) while T might be any.
		fnVal := reflect.ValueOf(sub.handler)
		args := []reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(event),
		}

		results := fnVal.Call(args)
		if !results[0].IsNil() {
			err := results[0].Interface().(error)
			if b.strategy == StopOnFirstError {
				return err
			}
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// EmitAsync dispatches an event to listeners in a separate goroutine.
// If b is nil, it uses the default global bus.
func EmitAsync[T any](ctx context.Context, b *Bus, event T) {
	if b == nil {
		b = defaultBus
	}
	go func() {
		_ = Emit(ctx, b, event)
	}()
}
