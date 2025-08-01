# Commandment - Operation Pattern Framework

A Go library implementing the Command/Query pattern with type-safe operation creation, service injection, and centralized logging.

## Features

- **Clean Command/Query Separation**: Distinct interfaces for read and write operations
- **Type-Safe Service Injection**: Automatic dependency injection using Go generics
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
┌─────────────┐    holds ref  ┌─────────────┐    calls    ┌─────────────┐
│   Invoker   │──────────────►│   Command   │────────────►│  Receiver   │
│(OperationBus)│               │   (Query)   │             │ (Service)   │
└─────────────┘               └─────────────┘             └─────────────┘
```

## Quick Start

### 1. Install the Library

```bash
go get github.com/davidlee/commandment/pkg/operation
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
import "github.com/davidlee/commandment/pkg/operation"

// Command for creating users (mutates state)
type CreateUserCommand struct {
    Params  CreateUserParams
    Service UserService
    Meta    operation.OperationMetadata
    Logger  operation.Logger
}

func (c *CreateUserCommand) Execute(ctx context.Context) (User, error) {
    return operation.ExecuteOperation(c, func() (User, error) {
        return c.Service.CreateUser(ctx, c.Params)
    })
}

func (c *CreateUserCommand) Metadata() operation.OperationMetadata {
    return c.Meta
}

func (c *CreateUserCommand) Descriptor() operation.OperationDescriptor {
    return operation.OperationDescriptor{
        Type:     "CreateUserCommand",
        Params:   c.Params,
        Metadata: c.Meta,
    }
}

func (c *CreateUserCommand) GetMetadata() *operation.OperationMetadata { return &c.Meta }
func (c *CreateUserCommand) GetLogger() operation.Logger               { return c.Logger }

// Query for getting users (read-only)
type GetUserQuery struct {
    Params  GetUserParams
    Service UserService
    Meta    operation.OperationMetadata
    Logger  operation.Logger
}

// Implement similar methods...
```

### 5. Create Domain-Specific Bus

```go
type UserBus struct {
    bus *operation.OperationBus
}

func NewUserBus(bus *operation.OperationBus) *UserBus {
    return &UserBus{bus: bus}
}

func (b *UserBus) NewCreateUserCommand(params CreateUserParams) (*CreateUserCommand, error) {
    return operation.CreateOperation[*CreateUserCommand](b.bus, params)
}

func (b *UserBus) NewGetUserQuery(params GetUserParams) (*GetUserQuery, error) {
    return operation.CreateOperation[*GetUserQuery](b.bus, params)
}
```

### 6. Setup and Usage

```go
// Setup framework
registry := operation.NewServiceRegistry()
operation.RegisterService[UserService](registry, myUserService)

logger := myLogger // implement operation.Logger interface
operationBus := operation.NewOperationBus(registry, logger)
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
- **`pkg/operation/`** - Core framework that users import
  - `operation.go` - Base interfaces and metadata
  - `bus.go` - Operation bus and creation logic  
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