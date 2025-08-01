# Commandment - Operation Pattern POC

A proof-of-concept implementation of the Command/Query Operation Pattern extracted from the design document.

## Architecture

### Core Components

- **Operation Pattern**: Unified base interface with separate Command and Query types
- **Operation Bus**: Centralized service management and operation execution  
- **Service Registry**: Type-safe service injection using generics
- **Execution Wrapper**: Common logging and metadata handling via callback pattern

### Key Features

- **Type Safety**: Compile-time safety with generic types
- **Clean Client API**: Specific methods like `NewDisplayNodeTreeCommand()` instead of generic syntax
- **Query-Only Access**: Restrict clients to read-only operations via `QueryInvoker` interface
- **Serialization**: Commands can be serialized/deserialized for audit trails and queuing
- **Minimal Duplication**: Callback pattern eliminates repetitive logging code
- **Observability**: Structured logging with operation correlation and timing

## Directory Structure

```
internal/
├── operation/           # Core operation pattern implementation
│   ├── operation.go     # Base interfaces and common functionality
│   ├── bus.go          # Operation bus and invoker interfaces  
│   ├── registry.go     # Service registry
│   ├── operations.go   # Concrete operation implementations
│   ├── params.go       # Parameter and result types
│   └── services.go     # Service interfaces
└── services/           # Mock service implementations
    └── mock_services.go

examples/
└── basic/
    └── main.go         # Demo showing all features
```

## Usage Example

```go
// Setup
registry := operation.NewServiceRegistry()
registry.Register[operation.TreeService](services.NewMockTreeService())
registry.Register[operation.ListService](services.NewMockListService())
registry.Register[operation.NodeService](services.NewMockNodeService())

bus := operation.NewOperationBus(registry, logger)

// Execute command
params := operation.DisplayNodeTreeCommandParams{
    RootReference: "root-123",
    MaxDepth:      3,
}

cmd, err := bus.NewDisplayNodeTreeCommand(params)
if err != nil {
    return err
}

result, err := cmd.Execute(context.Background())
```

## Running the Demo

```bash
cd examples/basic
go run main.go
```

## Design Benefits

1. **Maintainable**: Changes to logging/metadata have zero blast radius
2. **Type Safe**: Compile-time guarantees, no runtime type assertions needed
3. **Testable**: Easy to mock services and test operations in isolation
4. **Extensible**: Adding new operations requires minimal boilerplate
5. **Observable**: Rich structured logging with operation correlation
6. **Serializable**: Operations can be persisted for audit trails and replay

## Pattern Comparison

| Aspect | Before | After |
|--------|--------|--------|
| Logging Code | ~35 lines per operation | 3 lines + shared function |
| Type Safety | Runtime assertions | Compile-time generics |
| Client API | `CreateCommand[*T](bus, params)` | `bus.NewTCommand(params)` |
| Service Access | Direct coupling | Registry-based injection |
| Observability | Manual, inconsistent | Automatic, structured |
| Testing | Complex setup | Clean interfaces, easy mocking |