# Unit Testing Guide

Unit tests verify individual components in isolation. They form the foundation of your test suite.

## What Makes a Good Unit Test?

### Characteristics
- **Fast**: Milliseconds per test (< 100ms)
- **Isolated**: No dependencies on external systems
- **Focused**: Tests one specific behavior
- **Deterministic**: Same result every time
- **Self-contained**: Creates own test data

### Scope
A unit is typically:
- A single function/method
- A class with minimal dependencies
- A module with cohesive functionality

## Structure and Organization

### File Organization

Tests accompany code being unit tested, as per Go conventions.

### Test Naming

Be specific about what is tested and the expected outcome. It's ok if this leads to very long test function names.

## Isolation Techniques

### 1. Dependency Injection

What's good for the tests is good for the code.

### 2. Test Doubles

For both isolation and repeatable results.

### 3. Time Control

Remove determinism from tests by using static values or a library.

## Common Patterns

### Testing Error Conditions

Be descriptive in failure messages.

### Testing Side Effects

Consider spies / mocks to assert.

### Table-Driven Tests

```go
func TestStrings(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "empty string",
            input: "",
            want:  "",
        },
        {
            name:    "invalid input",
            input:   "invalid",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ProcessString(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ProcessString() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("ProcessString() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test Helpers

- Mark with `t.Helper()`
- Use descriptive names: `mustCreateUser`, `requireNoError`
- Keep tests running after failures when possible (`t.Error` over `t.Fatal`)

## Best Practices

### 1. One Assertion Per Test

Promotes SRP, good naming, and diagnostically useful failures.

### 2. Test Behavior, Not Implementation

Test the public interface and observable behaviour. Don't couple tests to
implementation details or internals.

### 3. Factories for Complex Objects

Use them in tests for consistency, brevity, and the avoidance of errors.

## Performance Considerations

- **Avoid I/O**: No file system, network, or database access
- **Minimize setup**: Keep arrangements simple
- **Parallel execution**: Design for concurrent test runs
- **Memory usage**: Clean up large test objects

## Summary

Unit tests are your first line of defense against bugs. Keep them:
- Fast enough to run constantly
- Isolated from external dependencies
- Focused on single behaviors
- Readable and maintainable

Remember: If unit tests are hard to write, your code's design likely needs improvement.

## Also Read

@GO_STYLE.md