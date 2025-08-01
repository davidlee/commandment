# Google Go Style Guide - Condensed Reference

Summary of idioms & anti-patterns from [Google style guide](https://google.github.io/styleguide/go/guide).

## Core Go Philosophy

The Google Go Style Guide emphasizes a **readability hierarchy** that
prioritizes code clarity above all else. Go code should be **clear**,
**simple**, **concise**, **maintainable**, and **consistent** - in that order.
The overarching principle is "clarity over cleverness": write for the reader,
not the compiler.

## Essential Naming Conventions

### Package Names
- **Always lowercase, single-word**: `tabwriter`, `creditcard` (never `tab_writer` or `creditCard`)
- **Singular, not plural**: `net/url` not `net/urls`
- **Avoid generic names**: no `util`, `common`, `helper`, `models`

### Variable Names
- **Length proportional to scope**: `i` for loop counter, `activeUserCount` for function parameter
- **Receiver names**: Short, consistent abbreviations (`u *User` throughout all methods)
- **No Get prefix** for getters: `user.Name()` not `user.GetName()`

### Interface Names
- **Method name + -er suffix**: `Reader`, `Writer`, `Stringer`
- **NO "I" prefix** (unlike C#/Java)
- Interfaces belong in the consuming package, not the implementing package

### Constants and Acronyms
- **MixedCaps**, never SCREAMING_SNAKE_CASE: `MaxPacketSize` not `MAX_PACKET_SIZE`
- **Consistent case for acronyms**: `XMLParser`, `HTTPSClient`, `userID`

## Code Organization

### Import Grouping
```go
import (
    // Standard library
    "fmt"
    "os"
    
    // Third party
    "github.com/pkg/errors"
    
    // Project packages
    "myproject/internal/auth"
)
```

### Declaration Order
1. Package comment
2. Package declaration
3. Imports (grouped as above)
4. Constants, variables, types, functions (logical grouping)

## Error Handling Patterns

### The Go Way
```go
// Always check and handle errors explicitly
if err := doSomething(); err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Define sentinel errors
var ErrNotFound = errors.New("item not found")

// Use custom error types for rich errors
type ValidationError struct {
    Field string
    Value string
}
```

### Error Wrapping Best Practices
- Use `%w` verb for wrapping: `fmt.Errorf("context: %w", err)`
- Only wrap when caller needs to unwrap
- Never ignore errors with `_`

## Documentation Standards

### Package Documentation
```go
// Package json implements encoding and decoding of JSON as defined in
// RFC 7159. The mapping between JSON and Go values is described
// in the documentation for the Marshal and Unmarshal functions.
package json
```

### Function Documentation
```go
// Marshal returns the JSON encoding of v.
//
// Marshal traverses the value v recursively...
func Marshal(v interface{}) ([]byte, error)
```

**Key principles**: Start with the name, use complete sentences, focus on what/why not how.

## Core Go Idioms

### Zero Values
```go
// Good - leverage zero values
var users []User      // nil slice is usable
var config Config     // struct with zero values
var mu sync.Mutex     // ready to use

// Bad - unnecessary initialization
var users = []User{}
```

### Multiple Return Values
```go
// The comma-ok idiom
value, ok := m[key]
if !ok {
    // handle missing key
}

// Error as last return value
result, err := function()
if err != nil {
    return nil, err
}
```

### Defer for Cleanup
```go
file, err := os.Open(filename)
if err != nil {
    return err
}
defer file.Close()  // Guaranteed cleanup

mu.Lock()
defer mu.Unlock()   // Pairs nicely with Lock
```

### Interface Satisfaction
```go
// Implicit satisfaction - no "implements" keyword
type Writer interface {
    Write([]byte) (int, error)
}

type FileWriter struct{}

func (fw *FileWriter) Write(data []byte) (int, error) {
    // FileWriter automatically satisfies Writer
}
```

## Performance Guidelines

### Key Principles
1. **Measure first** - use benchmarks and profiling
2. **Readability over micro-optimizations**
3. **Focus on algorithmic improvements**

### Memory Optimization
```go
// Pre-allocate slices when size is known
users := make([]User, 0, expectedCount)

// Use strings.Builder for concatenation
var b strings.Builder
for _, word := range words {
    b.WriteString(word)
}
```

## Concurrency Patterns

### Goroutine Management
```go
func (w *Worker) Run(ctx context.Context) error {
    var wg sync.WaitGroup
    
    for work := range w.workChan {
        wg.Add(1)
        go func(work Work) {
            defer wg.Done()
            w.process(ctx, work)
        }(work)
    }
    
    wg.Wait()
    return nil
}
```

### Channel Direction
```go
func produce(out chan<- int)    // Send-only
func consume(in <-chan int)     // Receive-only
```

### Context Usage
- **Always first parameter**: `func Process(ctx context.Context, data []byte)`
- **Propagate cancellation**: Check `ctx.Done()` in loops
- **Never store in structs**: Pass through function calls

## Package Design Principles

### API Design

- **Accept interfaces, return concrete types**
- **Small, focused packages** with single responsibility
- **Minimize dependencies**
- **Stable APIs** - avoid breaking changes

### Constructor Patterns
```go
// Simple constructor
func NewUser(name string) *User {
    return &User{
        Name:      name,
        CreatedAt: time.Now(),
    }
}

// Options pattern for complex construction
type UserOptions struct {
    Email string
    Admin bool
}

func NewUserWithOptions(name string, opts UserOptions) *User {
    return &User{
        Name:  name,
        Email: opts.Email,
        Admin: opts.Admin,
    }
}
```

## Summary

Go's style emphasizes **simplicity and clarity** over clever abstractions. The language deliberately omits many features common in other languages (inheritance, generics until recently, exceptions) in favor of explicit, straightforward code. When writing Go:

1. **Embrace simplicity** - don't recreate OOP patterns
2. **Be explicit** - especially with error handling
3. **Think in composition** - not inheritance
4. **Use the standard library** as your style guide
5. **Write for the reader** - your future self and teammates

The key to writing good Go is to work with the language, not against it. When in doubt, look at how the standard library solves similar problems - it exemplifies idiomatic Go.