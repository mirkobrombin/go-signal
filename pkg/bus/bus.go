package bus

import (
	"context"
	"errors"
	"reflect"
	"sort"

	"github.com/mirkobrombin/go-foundation/pkg/options"
	"github.com/mirkobrombin/go-foundation/pkg/safemap"
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
	subscribers *safemap.Map[reflect.Type, []subscriber]
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

// Option defines a functional configuration for the Bus.
type Option = options.Option[Bus]

// New creates a new Bus instance.
func New(opts ...Option) *Bus {
	b := &Bus{
		subscribers: safemap.New[reflect.Type, []subscriber](),
		strategy:    StopOnFirstError,
	}
	options.Apply(b, opts...)
	return b
}

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

	b.subscribers.Compute(key, func(subs []subscriber, exists bool) []subscriber {
		newSubs := append(subs, subscriber{
			handler:  fn,
			priority: p,
		})
		sort.SliceStable(newSubs, func(i, j int) bool {
			return newSubs[i].priority > newSubs[j].priority
		})
		return newSubs
	})
}

// Emit dispatches an event to all registered listeners synchronously.
// If b is nil, it uses the default global bus.
func Emit[T any](ctx context.Context, b *Bus, event T) error {
	if b == nil {
		b = defaultBus
	}

	// Use static type T to match Subscribe[T] key
	key := reflect.TypeFor[T]()

	subs, ok := b.subscribers.Get(key)
	if !ok {
		return nil
	}

	var errs []error
	for _, sub := range subs {
		// Direct type assertion (fast path)
		if fn, ok := sub.handler.(Handler[T]); ok {
			if err := fn(ctx, event); err != nil {
				if b.strategy == StopOnFirstError {
					return err
				}
				errs = append(errs, err)
			}
		} else {
			// Fallback or panic? Should never happen if T matches key.
			// But safemap is [reflect.Type, []subscriber].
			// subscriber.handler is 'any'.
			// If we retrieved by reflect.TypeFor[T], handler MUST be Handler[T].
			// Unless memory corruption or manual manipulation.
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
