package signal

import (
	"context"

	"github.com/mirkobrombin/go-signal/v2/pkg/bus"
)

type Bus = bus.Bus
type Handler[T any] = bus.Handler[T]
type Priority = bus.Priority
type DispatchStrategy = bus.DispatchStrategy

const (
	PriorityHigh   = bus.PriorityHigh
	PriorityNormal = bus.PriorityNormal
	PriorityLow    = bus.PriorityLow

	StopOnFirstError = bus.StopOnFirstError
	BestEffort       = bus.BestEffort
)

var (
	New          = bus.New
	Default      = bus.Default
	WithStrategy = bus.WithStrategy
)

func Subscribe[T any](b *Bus, fn Handler[T], priority ...Priority) {
	bus.Subscribe(b, fn, priority...)
}

func Emit[T any](ctx context.Context, b *Bus, event T) error {
	return bus.Emit(ctx, b, event)
}

func EmitAsync[T any](ctx context.Context, b *Bus, event T) {
	bus.EmitAsync(ctx, b, event)
}

func SubscribeWildcard(b *Bus, fn func(ctx context.Context, event any) error) {
	bus.SubscribeWildcard(b, fn)
}
