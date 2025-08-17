package commandment_test

import (
	"context"
	"testing"

	"github.com/davidlee/commandment/pkg/commandment"
)

// Test Logger implementation
type TestLogger struct{}

func (l *TestLogger) Info(msg string, keysAndValues ...any)  {}
func (l *TestLogger) Warn(msg string, keysAndValues ...any)  {}
func (l *TestLogger) Error(msg string, keysAndValues ...any) {}
func (l *TestLogger) Debug(msg string, keysAndValues ...any) {}

// Example service interface for testing
type TestService interface {
	DoSomething(ctx context.Context, input string) (string, error)
}

// Mock service implementation
type MockTestService struct{}

func (s *MockTestService) DoSomething(ctx context.Context, input string) (string, error) {
	return "result: " + input, nil
}

// Example operation for testing
type TestOperation struct {
	Params  string
	Service TestService
	Meta    commandment.OperationMetadata
	Logger  commandment.Logger
}

func (op *TestOperation) Execute(ctx context.Context) (string, error) {
	return commandment.ExecuteOperation(ctx, op, func(ctx context.Context) (string, error) {
		return op.Service.DoSomething(ctx, op.Params)
	})
}

func (op *TestOperation) Metadata() commandment.OperationMetadata {
	return op.Meta
}

func (op *TestOperation) Descriptor() commandment.OperationDescriptor {
	return commandment.OperationDescriptor{
		Type:     "TestOperation",
		Params:   op.Params,
		Metadata: op.Meta,
	}
}

func (op *TestOperation) GetMetadata() *commandment.OperationMetadata { return &op.Meta }
func (op *TestOperation) GetLogger() commandment.Logger               { return op.Logger }

func TestOperationFramework(t *testing.T) {
	// Setup
	registry := commandment.NewServiceRegistry()
	commandment.RegisterService[TestService](registry, &MockTestService{})

	logger := &TestLogger{}
	bus := commandment.NewOperationBus(registry, logger)

	// Test operation creation
	op, err := commandment.CreateOperation[*TestOperation](bus, "test input")
	if err != nil {
		t.Fatalf("Failed to create operation: %v", err)
	}

	if op == nil {
		t.Fatal("Operation should not be nil")
	}

	// Test execution
	result, err := op.Execute(context.Background())
	if err != nil {
		t.Fatalf("Operation execution failed: %v", err)
	}

	expected := "result: test input"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Test metadata
	metadata := op.Metadata()
	if metadata.UUID == "" {
		t.Error("Expected UUID to be set")
	}

	if metadata.Created.IsZero() {
		t.Error("Expected Created timestamp to be set")
	}
}

func TestServiceRegistry(t *testing.T) {
	registry := commandment.NewServiceRegistry()

	// Test registration and retrieval
	mockService := &MockTestService{}
	commandment.RegisterService[TestService](registry, mockService)

	retrieved := commandment.GetService[TestService](registry)
	if retrieved != mockService {
		t.Error("Retrieved service should be the same instance as registered")
	}
}
