# Operation Pattern Design (Command/Query)

## Design Goals

- Command pattern implementation which supports/separates commands and queries
- Decoupling of clients from service implementations 

- Proper Command pattern separation (Client → Invoker → Command → Receiver)
   - Centralized service management and middleware
   - Command reusability across CLI, TUI, and future (e.g. web) interfaces
   - Request/response logging and audit trails
   - Commands must be serializable / marshalable
   - Open/Closed principle: adding a command should not require updates to pattern infrastructure
   - Future potential for undo/redo, command queuing, distributed execution

Implement proper **Command Pattern separation** with four distinct architectural roles:

```
┌─────────────┐    creates    ┌─────────────┐
│   Client    │──────────────►│   Command   │
│ (CLI Parser)│               │ (ViceQuery) │
└─────────────┘               └─────────────┘
                                      │
                                      ▼
┌─────────────┐    holds ref  ┌─────────────┐    calls    ┌─────────────┐
│   Invoker   │──────────────►│   Command   │────────────►│  Receiver   │
│(CommandBus) │               │ (ViceQuery) │             │ (Service)   │
└─────────────┘               └─────────────┘             └─────────────┘
```

### Role Definitions

1. **Client (CLI/TUI CobraCommands)**: Parse user input, create Command objects, delegate to Invoker
   - Zero knowledge of services or infrastructure (includes queries! and repositories!)
   - Only knows about Command creation and result presentation

2. **Invoker (CommandBus/QueryBus)**: Centralize service management and command execution
   - Holds references to all services (Receivers)
   - Executes Commands by calling their Execute() methods with appropriate services
   - Handles cross-cutting concerns (logging, validation, middleware)

3. **Command (ViceCommand/ViceQuery)**: Encapsulate requests and execution target
   - Self-contained objects with request data and receiver "address"
   - Type-safe Execute() methods which take no arguments (except contex.Context)

4. **Receiver (Services)**: Perform actual business operations
   - Application services implementing business logic
   - No knowledge of how they're invoked (Command objects, CommandBus, etc) 
   - Return domain objects, or structs

### Core Architecture Rules

- **Clients** must have ZERO knowledge of services or infrastructure
- **Invoker** centralizes all service dependencies and command routing
- **Commands** are self-executing and reusable
- **Receivers** focus purely on business logic without UI concerns

## Critical Analysis & Concrete Design

### Original Design Issues

The initial design had several fundamental problems that required resolution:

1. **Service Discovery Undefined**: "Magic functions" indicated unresolved service-to-command mapping
2. **Type Safety Contradiction**: Promised compile-time safety but used generic interfaces requiring runtime resolution
3. **Command State Management**: Commands created with arguments but Execute() takes none - where do arguments live?
4. **Responsibility Confusion**: Bus holds services but commands execute themselves - who calls service methods?

## Concrete Implementation Design

#Note context7:get-library-docs (MCP)(context7CompatibleLibraryID: "/golang/go", topic: "generics method receivers")

### 1. Service Registry with Convention-Based Discovery

```go
type ServiceRegistry struct {
    services map[reflect.Type]interface{}
}

func (r *ServiceRegistry) Register[T any](service T) {
    r.services[reflect.TypeOf((*T)(nil)).Elem()] = service
}

func (r *ServiceRegistry) Get[T any]() T {
    serviceType := reflect.TypeOf((*T)(nil)).Elem()
    service, exists := r.services[serviceType]
    if !exists {
        panic(fmt.Sprintf("Service %T not registered", (*T)(nil)))
    }
    return service.(T)
}
```

### 2. Self-Contained Commands

```go
// Shared base interface for commands and queries
type Operation[TResult any] interface {
    Execute(ctx context.Context) (TResult, error)
    Metadata() OperationMetadata
    Descriptor() OperationDescriptor
}

// Commands mutate state
type Command[TResult any] interface {
    Operation[TResult]
}

// Queries are read-only
type Query[TResult any] interface {
    Operation[TResult]
}

// Concrete command implementation (updates node refs)
type DisplayNodeTreeCommand struct {
    Params        DisplayNodeTreeCommandParams
    Service       TreeService          // Injected during creation
    OperationMeta OperationMetadata    // Injected during creation
    Logger        Logger               // Injected during creation
}

// Concrete query implementation (read-only)
type ShowNodeQuery struct {
    Params        ShowNodeQueryParams
    Service       NodeService          // Injected during creation
    OperationMeta OperationMetadata    // Injected during creation
    Logger        Logger               // Injected during creation
}

func (q *ShowNodeQuery) Execute(ctx context.Context) (Node, error) {
    return executeOperation(q, func() (Node, error) {
        return q.Service.ShowNode(ctx, q.Params)
    })
}

func (q *ShowNodeQuery) Metadata() OperationMetadata {
    return q.OperationMeta
}

func (q *ShowNodeQuery) Descriptor() OperationDescriptor {
    return OperationDescriptor{
        Type:     "ShowNodeQuery",
        Params:   q.Params,
        Metadata: q.OperationMeta,
    }
}

func (q *ShowNodeQuery) getMetadata() *OperationMetadata { return &q.OperationMeta }
func (q *ShowNodeQuery) getLogger() Logger { return q.Logger }

func (c *DisplayNodeTreeCommand) Execute(ctx context.Context) (NodeTree, error) {
    return executeOperation(c, func() (NodeTree, error) {
        return c.Service.DisplayTree(ctx, c.Params)
    })
}

func (c *DisplayNodeTreeCommand) Metadata() OperationMetadata {
    return c.OperationMeta
}

func (c *DisplayNodeTreeCommand) Descriptor() OperationDescriptor {
    return OperationDescriptor{
        Type:     "DisplayNodeTreeCommand",
        Params:   c.Params,
        Metadata: c.OperationMeta,
    }
}

func (c *DisplayNodeTreeCommand) getMetadata() *OperationMetadata { return &c.OperationMeta }
func (c *DisplayNodeTreeCommand) getLogger() Logger { return c.Logger }

// Concrete command implementation (mutates state)
type CreateListCommand struct {
    Params        CreateListCommandParams
    Service       ListService
    OperationMeta OperationMetadata
    Logger        Logger
}

func (c *CreateListCommand) Execute(ctx context.Context) (NodeCommandResult, error) {
    return executeOperation(c, func() (NodeCommandResult, error) {
        return c.Service.CreateList(ctx, c.Params)
    })
}

func (c *CreateListCommand) Metadata() OperationMetadata {
    return c.OperationMeta
}

func (c *CreateListCommand) Descriptor() OperationDescriptor {
    return OperationDescriptor{
        Type:     "CreateListCommand", 
        Params:   c.Params,
        Metadata: c.OperationMeta,
    }
}

func (c *CreateListCommand) getMetadata() *OperationMetadata { return &c.OperationMeta }
func (c *CreateListCommand) getLogger() Logger { return c.Logger }
```

### 3. Operation Bus with Type-Safe Creation

```go
type OperationBus struct {
    registry *ServiceRegistry
    logger   Logger  // From internal/logging/logger.go
}

func NewOperationBus(registry *ServiceRegistry, logger Logger) *OperationBus {
    return &OperationBus{
        registry: registry,
        logger:   logger,
    }
}

// Query-only interface for read-only operations
type QueryInvoker interface {
    NewShowNodeQuery(params ShowNodeQueryParams) (*ShowNodeQuery, error)
}

// Command-only interface for mutations
type CommandInvoker interface {
    NewDisplayNodeTreeCommand(params DisplayNodeTreeCommandParams) (*DisplayNodeTreeCommand, error)
    NewCreateListCommand(params CreateListCommandParams) (*CreateListCommand, error)
}

// Full interface includes both queries and commands
type OperationInvoker interface {
    QueryInvoker
    CommandInvoker
}

// OperationBus implements OperationInvoker by delegating to generic method
func (b *OperationBus) NewShowNodeQuery(params ShowNodeQueryParams) (*ShowNodeQuery, error) {
    return CreateOperation[*ShowNodeQuery](b, params)
}

func (b *OperationBus) NewDisplayNodeTreeCommand(params DisplayNodeTreeCommandParams) (*DisplayNodeTreeCommand, error) {
    return CreateOperation[*DisplayNodeTreeCommand](b, params)
}

func (b *OperationBus) NewCreateListCommand(params CreateListCommandParams) (*CreateListCommand, error) {
    return CreateOperation[*CreateListCommand](b, params)
}

// Generic method - used internally by OperationInvoker methods
func CreateOperation[TOp Operation[TResult], TResult any](
    bus *OperationBus, 
    params any,
) (TOp, error) {
    // Use reflection to determine required service type
    serviceType := getRequiredServiceType[TOp]()
    service := bus.registry.Get(serviceType)
    
    // Create metadata for new command
    metadata := OperationMetadata{
        UUID:    generateUUID(),
        Created: time.Now(),
    }
    
    // Log operation creation
    opTypeName := reflect.TypeOf((*TOp)(nil)).Elem().Name()
    bus.logger.Info("Operation created",
        "operation_type", opTypeName,
        "operation_id", metadata.UUID,
        "service_type", serviceType.Name(),
    )
    
    // Create operation with injected service, metadata, and logger
    op, err := newOperationWithService[TOp](params, service, metadata, bus.logger)
    if err != nil {
        bus.logger.Error("Operation creation failed",
            "operation_type", opTypeName,
            "operation_id", metadata.UUID,
            "error", err,
        )
        return op, err
    }
    
    return op, nil
}

// Helper function to extract service type from operation type
func getRequiredServiceType[TOp any]() reflect.Type {
    opType := reflect.TypeOf((*TOp)(nil)).Elem()
    // Convention: look for service field
    serviceField, _ := opType.FieldByName("service")
    return serviceField.Type
}

// Factory function for operation creation
func newOperationWithService[TOp any](params any, service any, metadata OperationMetadata, logger Logger) (TOp, error) {
    // Use reflection to create operation instance with params, service, metadata, and logger
    opType := reflect.TypeOf((*TOp)(nil)).Elem()
    opValue := reflect.New(opType).Elem()
    
    opValue.FieldByName("Params").Set(reflect.ValueOf(params))
    opValue.FieldByName("Service").Set(reflect.ValueOf(service))
    opValue.FieldByName("OperationMeta").Set(reflect.ValueOf(metadata))
    opValue.FieldByName("Logger").Set(reflect.ValueOf(logger))
    
    return opValue.Addr().Interface().(TOp), nil
}

func generateUUID() string {
    // Implementation-specific UUID generation
    return uuid.New().String() // or whatever UUID library you prefer
}

// Generic execution wrapper that handles common logging and metadata
func executeOperation[T any](op operationWithMetadata, businessLogic func() (T, error)) (T, error) {
    op.getMetadata().Executed = time.Now()
    
    opTypeName := reflect.TypeOf(op).Elem().Name()
    logger := op.getLogger()
    metadata := op.getMetadata()
    
    logger.Info("Operation execution started",
        "operation_type", opTypeName,
        "operation_id", metadata.UUID,
    )
    
    result, err := businessLogic()
    op.getMetadata().Returned = time.Now()
    
    duration := op.getMetadata().Returned.Sub(op.getMetadata().Executed)
    if err != nil {
        logger.Error("Operation execution failed",
            "operation_type", opTypeName,
            "operation_id", metadata.UUID,
            "duration_ms", duration.Milliseconds(),
            "error", err,
        )
    } else {
        logger.Info("Operation execution completed",
            "operation_type", opTypeName,
            "operation_id", metadata.UUID,
            "duration_ms", duration.Milliseconds(),
        )
    }
    
    return result, err
}

// Helper interface for accessing operation metadata and logger
type operationWithMetadata interface {
    getMetadata() *OperationMetadata
    getLogger() Logger
}

// Deserialization: recreate executable operations from descriptors
func (b *OperationBus) CreateFromDescriptor(descriptor OperationDescriptor) (interface{}, error) {
    // Map command type to creation function
    switch descriptor.Type {
    case "DisplayNodeTreeCommand":
        params := DisplayNodeTreeCommandParams{}
        if err := json.Unmarshal(mustMarshal(descriptor.Params), &params); err != nil {
            return nil, err
        }
        return b.createDisplayNodeTreeCommand(params, descriptor.Metadata)
        
    case "CreateListCommand":
        params := CreateListCommandParams{}
        if err := json.Unmarshal(mustMarshal(descriptor.Params), &params); err != nil {
            return nil, err
        }
        return b.createCreateListCommand(params, descriptor.Metadata)
        
    default:
        return nil, fmt.Errorf("unknown command type: %s", descriptor.Type)
    }
}

func (b *OperationBus) createDisplayNodeTreeCommand(params DisplayNodeTreeCommandParams, metadata OperationMetadata) (*DisplayNodeTreeCommand, error) {
    service := GetService[TreeService](b.registry)
    return &DisplayNodeTreeCommand{
        Params:        params,
        Service:       service,
        OperationMeta: metadata,
        Logger:        b.logger,
    }, nil
}

func (b *OperationBus) createCreateListCommand(params CreateListCommandParams, metadata OperationMetadata) (*CreateListCommand, error) {
    service := GetService[ListService](b.registry)
    return &CreateListCommand{
        Params:        params,
        Service:       service,
        OperationMeta: metadata,
        Logger:        b.logger,
    }, nil
}
```

### 4. Client Usage Pattern

```go
// Clean client code using OperationInvoker interface
func ExecuteTreeDisplay(invoker OperationInvoker, rootRef string, maxDepth int) (NodeTree, error) {
    params := DisplayNodeTreeCommandParams{
        RootReference: rootRef,
        MaxDepth:      maxDepth,
    }
    
    cmd, err := invoker.NewDisplayNodeTreeCommand(params)
    if err != nil {
        return NodeTree{}, err
    }
    
    // Access metadata before/after execution
    fmt.Printf("Executing command %s created at %v\n", cmd.Metadata().UUID, cmd.Metadata().Created)
    
    result, err := cmd.Execute(context.Background())
    
    // Access execution timing
    fmt.Printf("Command completed in %v\n", cmd.Metadata().Returned.Sub(cmd.Metadata().Executed))
    
    return result, err
}

// Serialization example
func SerializeOperation(op Operation[any]) ([]byte, error) {
    descriptor := op.Descriptor()
    return json.Marshal(descriptor)
}

// Deserialization example  
func DeserializeAndExecute(bus *OperationBus, data []byte) (interface{}, error) {
    var descriptor OperationDescriptor
    if err := json.Unmarshal(data, &descriptor); err != nil {
        return nil, err
    }
    
    op, err := bus.CreateFromDescriptor(descriptor)
    if err != nil {
        return nil, err
    }
    
    // Type assertion needed since CreateFromDescriptor returns interface{}
    switch o := op.(type) {
    case *DisplayNodeTreeCommand:
        return o.Execute(context.Background())
    case *CreateListCommand:
        return o.Execute(context.Background())
    default:
        return nil, fmt.Errorf("unsupported operation type")
    }
}

func ExecuteListCreation(invoker OperationInvoker, title, description string) (NodeCommandResult, error) {
    params := CreateListCommandParams{
        Title:       title,
        Description: description,
        ParentID:    nil,
    }
    
    cmd, err := invoker.NewCreateListCommand(params)
    if err != nil {
        return NodeCommandResult{}, err
    }
    
    return cmd.Execute(context.Background())
}

// System initialization with logging
func InitializeOperationSystem(logger Logger) OperationInvoker {
    registry := &ServiceRegistry{services: make(map[reflect.Type]interface{})}
    
    // Register services
    registry.Register[TreeService](NewTreeService())
    registry.Register[ListService](NewListService())
    
    // Create operation bus with logger
    bus := NewOperationBus(registry, logger)
    
    return bus  // OperationBus implements OperationInvoker
}
```

### 5. Parameter and Result Structs

```go

// Client-safe structs with no dependencies

// DisplayNodeTreeCommand:
//
// This is a command because it has side effects
type DisplayNodeTreeCommandParams struct {
    RootReference string
    MaxDepth      int
}
// results can be, or contain, domain objects
type NodeTree struct {
    Nodes []Node
    Stats TreeStats
}

// CreateListCommand:
//
type CreateListCommandParams struct {
    Title       string
    Description string
    // ...
    ParentID    *int64
}
// returns a result type shared by several commands / queries 
type NodeCommandResult struct {
    Node     Node
    Errors   []ValidationError
}

// ShowNodeQuery:
//
type ShowNodeQueryParams struct {
    Ref      int64
    // ... 
}
// just returns (Node, err) so we don't need to define a result struct


// Optional metadata for audit trails and debugging
type OperationMetadata struct {
    UUID     string    `json:"uuid"`
    Created  time.Time `json:"created"`
    Executed time.Time `json:"executed,omitempty"`
    Returned time.Time `json:"returned,omitempty"`
}

// Serializable operation representation
type OperationDescriptor struct {
    Type     string            `json:"type"`
    Params   interface{}       `json:"params"`
    Metadata OperationMetadata `json:"metadata"`
}

// JSON serialization helper for type-safe params
func (od OperationDescriptor) MarshalJSON() ([]byte, error) {
    type Alias OperationDescriptor
    return json.Marshal(&struct {
        *Alias
        Params json.RawMessage `json:"params"`
    }{
        Alias:  (*Alias)(&od),
        Params: mustMarshal(od.Params),
    })
}

func mustMarshal(v interface{}) json.RawMessage {
    data, err := json.Marshal(v)
    if err != nil {
        panic(err) 
    }
    return data
}
```

### 6. Service Interface Definitions

```go
// Services remain focused on business logic
type TreeService interface {
    DisplayTree(ctx context.Context, params DisplayNodeTreeCommandParams) (NodeTree, error)
}

type ListService interface {
    CreateList(ctx context.Context, params CreateListCommandParams) (NodeCommandResult, error)
}

type NodeService interface {
    ShowNode(ctx context.Context, params ShowNodeQueryParams) (Node, error)
}

// Logger interface (from internal/logging/logger.go)
type Logger interface {
    Info(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
    Debug(msg string, keysAndValues ...interface{})
}
```

## Key Design Decision Points

### Resolved Decisions

1. **Service Discovery Strategy**: **Convention-based with explicit registration**
   - Commands embed service field names that map to registered service types
   - Bus uses reflection to match command types to services during creation
   - Balances type safety with implementation simplicity

1a. **Client API Design**: **OperationInvoker interface with specific methods**
   - Clean `invoker.NewDisplayNodeTreeCommand(params)` instead of generic `CreateOperation[T](bus, params)`
   - OperationBus implements OperationInvoker, hiding internal details from clients
   - Type-safe, IDE-friendly, easily mockable for testing

1b. **Naming and Type Structure**: **Operation as shared base interface**
   - `Operation[TResult]` as shared base interface for common behavior
   - Separate `Command[TResult]` and `Query[TResult]` interfaces extend Operation
   - `OperationBus` and `OperationInvoker` reveal intent better than generic "Bus"
   - `OperationDescriptor` for consistent serialization naming

2. **Type Safety Approach**: **Hybrid compile-time/runtime**
   - Operation creation is compile-time safe with specific methods
   - Bus execution uses runtime dispatch for flexibility
   - Provides both developer experience and system flexibility

3. **Command/Query Separation**: **Separate interfaces with shared base**
   - `Command` and `Query` as distinct interfaces extending `Operation`
   - Clear semantic distinction at the type level
   - Enables query-only or command-only interfaces for restricted access
   - Same execution pattern, different types for different intents

4. **Command Lifecycle**: **Self-contained commands**
   - Commands hold both parameters and injected service references
   - Bus injects service during command creation
   - Execute() method truly parameter-less as originally desired

5. **Error Handling**: **Explicit error returns**
   - Commands return (Result, error) tuples
   - Bus creation can fail if services not registered
   - Clear error propagation from service layer

### Open Decisions Requiring Future Resolution

6. **Middleware/Cross-Cutting Concerns**: **PARTIALLY RESOLVED**
   - **Logging**: Implemented via logger injection into commands
   - **Authorization**: Not applicable for this use case (single-user/local execution)
   - **Validation**: Handled by services returning errors for invalid inputs
   - **Open**: Should additional middleware wrap command execution for future needs?

7. **Asynchronous Execution**: **OPEN**
   - Should commands support async execution patterns?
   - How would results be delivered (channels, callbacks, futures)?
   - Impact on client code simplicity vs. system scalability

8. **Command Serialization**: **RESOLVED**
   - Commands provide serializable descriptors containing type + params + metadata
   - Service references excluded from serialization (runtime-only)
   - Bus can recreate executable commands from descriptors

9. **Transaction Management**: **OPEN**
   - How should transactional boundaries be managed?
   - Should commands participate in larger transactions?
   - Consider Unit of Work pattern integration

10. **Command Composition**: **OPEN**
    - Should commands be composable (command containing other commands)?
    - How to handle complex workflows spanning multiple services?
    - Consider saga pattern for distributed operations

11. **Performance Optimization**: **OPEN**
    - Reflection overhead in command creation acceptable?
    - Consider code generation vs. runtime reflection trade-offs
    - Pool command objects vs. create-per-use

12. **Testing Strategy**: **OPEN**
    - How to test commands in isolation from services?
    - Mock strategy for service dependencies
    - Integration test patterns for full command execution

### Implementation Notes

- **Field Naming**: The metadata field is named `OperationMeta` (not `Metadata`) to avoid Go's field/method name conflict with the `Metadata()` method
- **Naming Convention**: Commands follow `<Operation><Entity>Command` pattern
- **Client Interface**: Use `CommandInvoker` interface with specific `New*Command()` methods
  - Avoids generic syntax in client code
  - Provides clean, type-safe API
  - Easy to mock for testing
- **Required Fields**: All commands must have:
  - `Params` field for request data
  - `Service` field for service injection
  - `OperationMeta` field for operation tracking
  - `Logger` field for structured logging
- **Required Methods**: All commands must implement:
  - `Execute(ctx context.Context) (TResult, error)`
  - `Metadata() OperationMetadata`
  - `Descriptor() CommandDescriptor`
- **Interface Evolution**: Adding new commands requires extending `CommandInvoker` interface
  - Trade-off: explicit methods vs. generic flexibility
  - Benefit: compile-time safety and discoverability
- **Metadata Lifecycle**: 
  - `Created` timestamp set during `CreateCommand`
  - `Executed` timestamp set at start of `Execute()`
  - `Returned` timestamp set at end of `Execute()`
  - `UUID` generated during command creation for tracing
- **Serialization Strategy**:
  - Commands serialize to `CommandDescriptor` (type + params + metadata)
  - Service references excluded from serialization (runtime-only)
  - `CreateFromDescriptor()` recreates executable commands with fresh service injection
- **Logging Integration**:
  - Bus initialized with logger: `NewBus(registry, logger)`
  - Commands log creation, execution start/end, and errors
  - Structured logging with command_id correlation across lifecycle
  - Execution timing and business metrics captured
- **Context Propagation**: All Execute methods take context.Context for cancellation/tracing

### Serialization Use Cases

- **Audit Logging**: Serialize descriptors for command history
- **Command Queuing**: Store descriptors, execute later via deserialization  
- **Undo/Redo**: Persist descriptors for replay capability
- **Distributed Systems**: Send descriptors across network boundaries

## Known Limitations

### Asynchronous Execution

**Current Design**: Commands use synchronous `Execute(ctx) (TResult, error)` pattern that blocks until completion.

**Limitations**:
- No progress reporting for long-running operations
- No streaming results or partial updates
- Difficult to compose multiple async operations
- Poor cancellation granularity beyond context timeout
- No built-in retry or circuit breaker patterns

**Potential Solutions**:
- Future-based execution: `Execute() -> Future[TResult]`
- Event-driven model with progress channels
- Streaming interface for long-running commands
- Separate command submission from result retrieval

### Network Distribution

**Current Design**: Commands hold direct service interface references, assuming single-process execution.

**Limitations**:
- Service references cannot cross network boundaries
- No service discovery or load balancing mechanisms
- Results must be fully serializable (TResult cannot contain interfaces)
- No network-specific error handling (timeouts, retries, partitions)
- Command routing hardcoded to local service registry

**Potential Solutions**:
- Abstract services behind RPC client interfaces
- Implement distributed service registry with health checking
- Add command routing/sharding logic to bus
- Rich error types for network failures
- Separate local vs. remote execution paths

### Performance and Scalability

**Current Design**: Uses reflection for command creation and service injection.

**Limitations**:
- Reflection overhead in command factory methods
- No command pooling or reuse strategies
- Type switches required for deserialization
- All services loaded into single registry (memory usage)
- No lazy loading or service lifecycle management

**Potential Solutions**:
- Code generation alternative to reflection
- Command object pooling for high-frequency operations
- Lazy service instantiation and disposal
- Service partitioning by domain boundaries

### Type Safety Trade-offs

**Current Design**: Hybrid compile-time/runtime type approach.

**Limitations**:
- `CreateFromDescriptor()` returns `interface{}` requiring type assertions
- Service injection relies on reflection and naming conventions
- Parameter deserialization can fail at runtime despite compile-time safety
- Generic constraints become complex for command composition

**Potential Solutions**:
- Registry of typed factory functions instead of reflection
- Stronger compile-time guarantees for descriptor roundtrip
- Consider code generation for type-safe deserialization

### Testing Complexity

**Current Design**: Commands encapsulate service dependencies internally.

**Limitations**:
- Testing requires full bus and service registry setup
- Difficult to mock individual service interactions
- Command serialization testing requires JSON roundtrips
- Integration test complexity scales with service dependencies

**Potential Solutions**:
- Constructor injection alternative for test scenarios
- Mock service registry for isolated testing
- Command behavior testing separate from service integration

### Framework Lock-in

**Current Design**: Commands must follow specific struct patterns and naming conventions.

**Limitations**:
- All commands must have `params`, `service`, `metadata` fields
- Reflection depends on field names matching exactly
- Adding new required methods breaks existing commands
- Difficult to migrate existing code to command pattern

**Potential Solutions**:
- Interface-based dependency injection
- Builder pattern for command construction
- Gradual migration strategies for existing services