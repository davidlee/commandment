# Commandment - Operation Pattern Framework

A Go library implementing the Command/Query pattern with type-safe operation creation, service injection, centralized logging, and flexible dependency management.

## Features

- **Clean Command/Query Separation**: Distinct interfaces for read and write operations
- **Type-Safe Service Injection**: Automatic dependency injection using Go generics
- **Flexible Dependencies Access**: Optional Dependencies injection with context-based access
- **Context-Enriched Execution**: Operations receive enriched context with metadata and Dependencies
- **Centralized Logging**: Built-in operation lifecycle logging with structured data
- **Serializable Operations**: Support for operation persistence and reconstruction
- **Framework/Domain Separation**: Reusable framework with user-defined domain logic

## Architecture

The framework follows the Command Pattern with four distinct roles:

```
┌─────────────┐    creates    ┌─────────────┐
│   Client    │──────────────►│   Command   │
│ (CLI Parser)│               │   (Query)   │
└─────────────┘               └─────────────┘
                                      │
                                      ▼
┌──────────────┐    holds ref  ┌─────────────┐    calls    ┌─────────────┐
│   Invoker    │──────────────►│   Command   │────────────►│  Receiver   │
│(OperationBus)│               │   (Query)   │             │ (Service)   │
└──────────────┘               └─────────────┘             └─────────────┘
```

## Quick Start

### 1. Install the Library

```bash
go get github.com/davidlee/commandment/pkg/commandment
```

### 2. Define Your Domain Services

```go
package myapp

import "context"

// Define your service interfaces
type UserService interface {
    CreateUser(ctx context.Context, params CreateUserParams) (User, error)
    GetUser(ctx context.Context, params GetUserParams) (User, error)
}
```

### 3. Define Parameters and Domain Objects

```go
// Parameters for operations
type CreateUserParams struct {
    Name  string
    Email string
}

type GetUserParams struct {
    ID int64
}

// Domain objects
type User struct {
    ID    int64
    Name  string
    Email string
}
```

### 4. Implement Operations

```go
import "github.com/davidlee/commandment/pkg/commandment"

// Command for creating users (mutates state)
type CreateUserCommand struct {
    Params  CreateUserParams
    Service UserService
    Meta    commandment.OperationMetadata
    Logger  commandment.Logger
}

func (c *CreateUserCommand) Execute(ctx context.Context) (User, error) {
    return commandment.ExecuteOperation(ctx, c, func(ctx context.Context) (User, error) {
        return c.Service.CreateUser(ctx, c.Params)
    })
}

func (c *CreateUserCommand) Metadata() commandment.OperationMetadata {
    return c.Meta
}

func (c *CreateUserCommand) Descriptor() commandment.OperationDescriptor {
    return commandment.OperationDescriptor{
        Type:     "CreateUserCommand",
        Params:   c.Params,
        Metadata: c.Meta,
    }
}

func (c *CreateUserCommand) GetMetadata() *commandment.OperationMetadata { return &c.Meta }
func (c *CreateUserCommand) GetLogger() commandment.Logger               { return c.Logger }

// Query for getting users (read-only)
type GetUserQuery struct {
    Params  GetUserParams
    Service UserService
    Meta    commandment.OperationMetadata
    Logger  commandment.Logger
}

// Implement similar methods...
```

### 5. Create Domain-Specific Bus

```go
type UserBus struct {
    bus *commandment.OperationBus
}

func NewUserBus(bus *commandment.OperationBus) *UserBus {
    return &UserBus{bus: bus}
}

func (b *UserBus) NewCreateUserCommand(params CreateUserParams) (*CreateUserCommand, error) {
    return commandment.CreateOperation[*CreateUserCommand](b.bus, params)
}

func (b *UserBus) NewGetUserQuery(params GetUserParams) (*GetUserQuery, error) {
    return commandment.CreateOperation[*GetUserQuery](b.bus, params)
}
```

### 6. Setup and Usage

```go
// Setup framework
registry := commandment.NewServiceRegistry()
commandment.RegisterService[UserService](registry, myUserService)

logger := myLogger // implement commandment.Logger interface
operationBus := commandment.NewOperationBus(registry, logger)
userBus := NewUserBus(operationBus)

// Use operations
cmd, err := userBus.NewCreateUserCommand(CreateUserParams{
    Name:  "John Doe",
    Email: "john@example.com",
})
if err != nil {
    log.Fatal(err)
}

user, err := cmd.Execute(context.Background())
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created user: %+v\n", user)
```

## Examples

See the `/examples` directory for complete working examples:

- **`examples/nodemanager/`** - Complete domain implementation for node/tree management
- **`examples/basic/`** - Simple demo showing framework usage

Run the demo:
```bash
just demo
# or
cd examples/basic && go run main.go
```

## Package Structure

### Library (Reusable Framework)
- **`pkg/commandment/`** - Core framework that users import
  - `operation.go` - Base interfaces, metadata, and context enrichment
  - `bus.go` - Operation bus, creation logic, and Dependencies management
  - `registry.go` - Service registry with type-safe injection

### User Code (Domain-Specific)
- **Services** - Define your business service interfaces
- **Operations** - Implement concrete commands/queries
- **Parameters** - Define parameter structs and domain objects
- **Invokers** - Create domain-specific operation buses

## Key Concepts

### Operations vs Services
- **Operations** encapsulate a single unit of work with logging and metadata
- **Services** contain the actual business logic
- Operations use services but add framework capabilities

### Commands vs Queries
- **Commands** mutate state (implement `Command[T]` interface)
- **Queries** are read-only (implement `Query[T]` interface)
- Both share common `Operation[T]` behavior

### Service Injection
- Services are registered once in the `ServiceRegistry`
- Operations get services injected automatically during creation
- Type-safe service discovery using Go generics

### Metadata and Logging
- Every operation gets UUID, timestamps, and structured logging
- Operations log creation, execution start/end, and errors
- Full audit trail for all operations

## Advanced Features

### Context-Enriched Execution

Operations receive an enriched context during execution that includes operation metadata and optional Dependencies:

```go
func (c *CreateUserCommand) Execute(ctx context.Context) (User, error) {
    return commandment.ExecuteOperation(ctx, c, func(enrichedCtx context.Context) (User, error) {
        // Access operation metadata from context
        metadata := commandment.OperationMetadataFromContext(enrichedCtx)
        if metadata != nil {
            c.Logger.Info("Operation started", "operation_id", metadata.UUID)
        }
        
        // Your business logic with enriched context
        return c.Service.CreateUser(enrichedCtx, c.Params)
    })
}
```

### Dependencies Management

The framework supports flexible Dependencies injection for complex infrastructure needs:

#### 1. Define Your Dependencies

```go
// Define your application's Dependencies
type MyDependencies struct {
    db         *sql.DB
    eventStore EventStore
    logger     Logger
}

func (d *MyDependencies) NodeRepository() NodeRepository {
    return NewNodeRepository(d.db, d.logger)
}

func (d *MyDependencies) WithTransaction(fn func(*MyDependencies) error) error {
    tx, err := d.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    txDeps := &MyDependencies{db: tx, eventStore: d.eventStore, logger: d.logger}
    err = fn(txDeps)
    if err != nil {
        return err
    }
    return tx.Commit()
}
```

#### 2. Setup Bus with Default Dependencies

```go
// Initialize your Dependencies
deps := &MyDependencies{
    db:         initDB(),
    eventStore: initEventStore(),
    logger:     initLogger(),
}

// Create bus with default Dependencies
registry := commandment.NewServiceRegistry()
commandment.RegisterService[UserService](registry, myUserService)

bus := commandment.NewOperationBusWithDefaultDependencies(registry, logger, deps)
```

#### 3. Access Dependencies in Operations

**Context-Based Access (Recommended):**
```go
func (c *CreateUserCommand) Execute(ctx context.Context) (User, error) {
    return commandment.ExecuteOperation(ctx, c, func(ctx context.Context) (User, error) {
        // Access Dependencies from enriched context
        deps := commandment.DependenciesFromContext(ctx).(*MyDependencies)
        
        return deps.WithTransaction(func(txDeps *MyDependencies) error {
            repo := txDeps.NodeRepository()
            eventWriter := txDeps.EventWriter()
            
            // Complex operations with direct Dependencies access
            return c.createUserWithEvents(repo, eventWriter, c.Params)
        })
    })
}
```

**Direct Access:**
```go
func (c *CreateUserCommand) Execute(ctx context.Context) (User, error) {
    // Access Dependencies directly from operation
    deps := commandment.GetDependencies(c).(*MyDependencies)
    
    return commandment.ExecuteOperation(ctx, c, func(ctx context.Context) (User, error) {
        return deps.WithTransaction(func(txDeps *MyDependencies) error {
            // Use Dependencies for infrastructure concerns
            return c.Service.CreateUser(ctx, c.Params)
        })
    })
}
```

#### 4. Per-Operation Dependencies Override

```go
// Special Dependencies for specific operations
migrationDeps := &MigrationDependencies{
    sourceDB: sourceDB,
    targetDB: targetDB,
}

// Create operation with specific Dependencies
op, err := commandment.CreateOperationWithDependencies[*MigrateUsersCommand](
    bus, 
    migrationParams, 
    migrationDeps,
)
```

### Usage Patterns

The framework supports multiple patterns for different use cases:

**Pattern 1: Service-Only (Simple)**
- Use service injection for domain logic
- No Dependencies needed for simple operations

**Pattern 2: Direct Dependencies (Complex Infrastructure)**
- Access Dependencies directly for transaction management
- Ideal for bulk operations, migrations, complex queries

**Pattern 3: Hybrid (Flexible)**
- Use Dependencies for infrastructure (transactions, caching)
- Use Services for domain logic
- Best of both approaches

## Development

```bash
# Run tests
just test

# Run linter
just lint

# Run demo
just demo

# Build demo binary
just build
```

## Pattern Benefits

1. **Maintainable**: Changes to logging/metadata have zero blast radius
2. **Type Safe**: Compile-time guarantees, no runtime type assertions needed
3. **Testable**: Easy to mock services and test operations in isolation
4. **Extensible**: Adding new operations requires minimal boilerplate
5. **Observable**: Rich structured logging with operation correlation
6. **Publishable**: Clean separation between framework and domain code

## License

MIT License - see LICENSE file for details.
