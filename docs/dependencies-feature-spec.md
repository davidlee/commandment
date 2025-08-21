# Dependencies Access Feature Specification

## Overview

This specification defines a feature for the commandment library that provides easy access to user-defined Dependencies objects from operation execution contexts, without making the API more cumbersome for normal usage.

## Goals

1. **Easy Access**: Operations can access a Dependencies object when needed
2. **Non-Intrusive**: Normal usage patterns remain unchanged
3. **No Import Coupling**: commandment library doesn't import user Dependencies types
4. **Flexible**: Support both default Dependencies and per-operation overrides
5. **Type Safe**: Compile-time safety where possible, runtime safety otherwise

## Current Architecture Analysis

### Existing Patterns
- Operations have exactly one `Service` field injected by type via reflection
- Context is enriched with operation metadata during execution
- Service instances are managed by `ServiceRegistry`
- No breaking changes allowed to existing operation patterns

### Import Coupling Analysis

**Operation Files**: Must import `MyDependencies` type for type assertion
```go
// operation.go - imports MyDependencies
deps := commandment.DependenciesFromContext(ctx).(*MyDependencies)
```

**Service Files**: Only import what they use from Dependencies
```go
// service.go - no MyDependencies import needed
func (s *NodeService) CreateNode(repo NodeRepository, params CreateParams) error {
    // Uses injected components, not full Dependencies
}
```

## API Design

### Core Functions

```go
// Context-based access (primary API)
func DependenciesFromContext(ctx context.Context) any

// Alternative: Operation-based access  
func GetDependencies(op OperationWithMetadata) any

// Bus constructor with default Dependencies
func NewOperationBusWithDefaultDependencies(registry *ServiceRegistry, logger Logger, defaultDeps any) *OperationBus

// Operation creation with Dependencies override
func CreateOperationWithDependencies[TOp Operation[TResult], TResult any](
    bus *OperationBus,
    params any,
    deps any,
) (TOp, error)
```

### Context Integration

Dependencies are stored in the execution context using the same enrichment pattern as operation metadata:

```go
// Internal context key (private)
const dependenciesKey contextKey = "commandment:dependencies"

// Internal helper for context enrichment
func WithDependencies(ctx context.Context, deps any) context.Context
```

## Usage Patterns

### Pattern 1: Default Dependencies (Most Common)

**Setup**:
```go
// Application bootstrap
deps := &MyDependencies{db: db, logger: logger}
registry := commandment.NewServiceRegistry()
bus := commandment.NewOperationBusWithDefaultDependencies(registry, logger, deps)
```

**Operation Implementation**:
```go
type CreateNodeCommand struct {
    Params  CreateNodeParams
    Service NodeService              // Still uses service injection
    Meta    commandment.OperationMetadata
    Logger  commandment.Logger
}

func (cmd *CreateNodeCommand) Execute(ctx context.Context) (NodeResult, error) {
    return commandment.ExecuteOperation(ctx, cmd, func(ctx context.Context) (NodeResult, error) {
        // Access Dependencies when service wrapper adds overhead
        deps := commandment.DependenciesFromContext(ctx).(*MyDependencies)
        
        return deps.WithTransaction(func(txDeps *MyDependencies) error {
            repo := txDeps.NodeRepository()
            eventWriter := txDeps.EventWriter()
            
            // Direct Dependencies usage for complex operations
            return cmd.createNodeWithEvents(repo, eventWriter, cmd.Params)
        })
        
        // Alternative: Still use service for simple operations
        // return cmd.Service.CreateNode(cmd.Params)
    })
}
```

### Pattern 2: Per-Operation Dependencies Override

**Special Dependencies**:
```go
// Different Dependencies type for migrations
type MigrationDependencies struct {
    sourceDB *sql.DB
    targetDB *sql.DB
    logger   Logger
}

// Create operation with specific Dependencies
migrationDeps := &MigrationDependencies{/*...*/}
op, err := commandment.CreateOperationWithDependencies(bus, params, migrationDeps)
```

**Operation Usage**:
```go
func (cmd *MigrateDataCommand) Execute(ctx context.Context) (MigrationResult, error) {
    return commandment.ExecuteOperation(ctx, cmd, func(ctx context.Context) (MigrationResult, error) {
        // Gets MigrationDependencies, not default Dependencies
        deps := commandment.DependenciesFromContext(ctx).(*MigrationDependencies)
        
        return cmd.migrateData(deps.sourceDB, deps.targetDB)
    })
}
```

### Pattern 3: Service + Dependencies Hybrid

**Service for Domain Logic**:
```go
type NodeService struct {
    // Injected domain components
}

func (s *NodeService) CreateNode(repo NodeRepository, params CreateParams) error {
    // Pure domain logic using injected repositories
}
```

**Operation for Infrastructure**:
```go
func (cmd *CreateNodeCommand) Execute(ctx context.Context) (NodeResult, error) {
    return commandment.ExecuteOperation(ctx, cmd, func(ctx context.Context) (NodeResult, error) {
        deps := commandment.DependenciesFromContext(ctx).(*MyDependencies)
        
        // Use Dependencies for transaction management
        return deps.WithTransaction(func(txDeps *MyDependencies) error {
            repo := txDeps.NodeRepository()
            
            // Use Service for business logic
            return cmd.Service.CreateNode(repo, cmd.Params)
        })
    })
}
```

## Implementation Strategy

### Phase 1: Context Integration

1. Add Dependencies context key and helper functions
2. Extend `ExecuteOperation` to enrich context with Dependencies
3. Implement `DependenciesFromContext` function

### Phase 2: Bus Enhancement

1. Add optional Dependencies field to `OperationBus`
2. Implement `NewOperationBusWithDefaultDependencies` constructor
3. Modify operation creation to include Dependencies in context

### Phase 3: Per-Operation Override

1. Implement `CreateOperationWithDependencies` function
2. Ensure per-operation Dependencies take precedence over default
3. Add validation and error handling

## Technical Details

### Context Storage

```go
type contextKey string
const dependenciesKey contextKey = "commandment:dependencies"

func WithDependencies(ctx context.Context, deps any) context.Context {
    return context.WithValue(ctx, dependenciesKey, deps)
}

func DependenciesFromContext(ctx context.Context) any {
    return ctx.Value(dependenciesKey)
}
```

### Bus Structure

```go
type OperationBus struct {
    registry    *ServiceRegistry
    logger      Logger
    defaultDeps any  // Optional default Dependencies
}
```

### Operation Creation Flow

1. `CreateOperation` or `CreateOperationWithDependencies` called
2. Dependencies (default or override) determined
3. Operation created with standard service injection
4. During `ExecuteOperation`, context enriched with Dependencies
5. Business logic accesses Dependencies via `DependenciesFromContext`

## Error Handling

### Missing Dependencies
```go
func DependenciesFromContext(ctx context.Context) any {
    deps := ctx.Value(dependenciesKey)
    if deps == nil {
        // Return nil - let caller handle missing Dependencies
        return nil
    }
    return deps
}
```

### Type Assertion Failures
Consumer responsibility:
```go
deps := commandment.DependenciesFromContext(ctx)
if deps == nil {
    return fmt.Errorf("no dependencies available")
}

myDeps, ok := deps.(*MyDependencies)
if !ok {
    return fmt.Errorf("unexpected dependencies type: %T", deps)
}
```

## Backward Compatibility

### Existing Operations
- No changes required to existing operation implementations
- Service injection continues to work unchanged
- Dependencies access is purely additive

### Existing Bus Usage
- `NewOperationBus` continues to work without Dependencies
- `CreateOperation` works with or without default Dependencies
- No breaking changes to any existing APIs

## Testing Strategy

### Unit Tests
- Context enrichment and retrieval
- Default Dependencies injection
- Per-operation Dependencies override
- Type safety and error handling

### Integration Tests
- Full operation lifecycle with Dependencies
- Service + Dependencies hybrid usage
- Multiple Dependencies types in same application

### Example Test Cases
```go
func TestDependenciesAccess(t *testing.T) {
    deps := &TestDependencies{value: "test"}
    bus := NewOperationBusWithDefaultDependencies(registry, logger, deps)
    
    op, err := CreateOperation[*TestOperation](bus, "params")
    require.NoError(t, err)
    
    result, err := op.Execute(context.Background())
    require.NoError(t, err)
    
    // Operation should have accessed Dependencies
    assert.Equal(t, "test-processed", result)
}
```

## Future Enhancements

### Optional Type-Safe API
If import coupling is acceptable in specific contexts:
```go
func GetTypedDependencies[T any](op OperationWithMetadata) (T, bool) {
    deps := GetDependencies(op)
    if deps == nil {
        var zero T
        return zero, false
    }
    
    typed, ok := deps.(T)
    return typed, ok
}
```

### Dependencies Validation
Runtime validation that Dependencies implement expected interfaces:
```go
func ValidateDependencies(deps any, requiredInterfaces ...reflect.Type) error {
    // Runtime interface checking
}
```

## Summary

This feature provides powerful Dependencies access while maintaining the simplicity and backward compatibility of the existing commandment library. It enables both the Service Wrapper pattern from T133 and direct Dependencies access for complex operations, giving users maximum flexibility in their architecture choices.

The context-based approach aligns with Go idioms and eliminates import coupling between the library and user Dependencies types, while the hybrid default/override system accommodates both simple and complex use cases.