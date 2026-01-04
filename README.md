# Go Signal

**Go Signal** is a high-performance, type-safe in-memory event bus for Go.
It leverages Go generics to provide a zero-boilerplate publish/subscribe mechanism within a single application process.

## Features

*   **Type Safe**: Events are typed structs. No `interface{}` casting or magic strings needed.
*   **Generics Based**: `Subscribe[UserCreated](...)` automatically infers the event type.
*   **Prioritized Listeners**: Control execution order with `PriorityHigh`, `PriorityNormal`, `PriorityLow`.
*   **Middleware Support**: Add logging, tracing, or error handling to the bus pipeline.
*   **Error Strategies**: Choose between `StopOnFirstError` or `BestEffort` execution.

## Installation

```bash
go get github.com/mirkobrombin/go-signal/v2
```

## Quick Start
See [examples/basic/main.go](examples/basic/main.go) for a runnable demo.

```go
type UserCreated struct {
    ID int
}

// Subscribe
bus.Subscribe(b, func(ctx context.Context, e UserCreated) error {
    fmt.Println("User Created:", e.ID)
    return nil
})

// Emit
bus.Emit(ctx, b, UserCreated{ID: 1})
```

## Documentation

*   [Getting Started](docs/getting-started.md)
*   [Architecture](docs/architecture.md)
