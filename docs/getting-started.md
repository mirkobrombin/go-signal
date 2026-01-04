# Getting Started

## 1. Create a Bus instance
You can configure the bus with options like execution strategy or middleware.

```go
import "github.com/mirkobrombin/go-signal/pkg/bus"

// Create a new bus instance
b := bus.New(
    bus.WithStrategy(bus.BestEffort),
)
```

## 2. Define Events
Events are simple Go structs.

```go
type OrderPlaced struct {
    OrderID string
    Amount  float64
}
```

## 3. Subscribe
Use the generic `Subscribe` function. The type argument is inferred from the handler signature.

```go
bus.Subscribe(b, func(ctx context.Context, event OrderPlaced) error {
    log.Printf("Order %s placed for %.2f", event.OrderID, event.Amount)
    return nil
})
```

## 4. Emit
Publish events using `Emit`.

```go
err := bus.Emit(context.Background(), b, OrderPlaced{
    OrderID: "123",
    Amount:  99.99,
})
```
