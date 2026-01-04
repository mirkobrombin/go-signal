# Architecture

## Design Philosophy

Go Signal is designed to decouple components within a **monolithic** or **modular** application.
It acts as a process-local event bus. For distributed messaging, use **Go Relay**.

## Internal Handling
The Bus maintains a map of subscribers indexed by the `reflect.Type` of the event struct.
When `Emit` is called:

1.  The event type is determined.
2.  Registered subscribers for that type are retrieved.
3.  Subscribers are sorted by Priority.
4.  Handlers are executed sequentially (synchronous execution).

## Synchronous by default
Go Signal is synchronous. When `Emit` returns, all handlers have been executed.
This makes it safe for use cases like transaction management or immediate validation.
If you need async processing, simply launch a goroutine inside your handler.
