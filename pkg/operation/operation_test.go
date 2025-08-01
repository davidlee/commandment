package operation_test

import (
	"context"
	"testing"

	"com.github/davidlee/commandment/pkg/operation"
)

// Test Logger implementation
type TestLogger struct{}

func (l *TestLogger) Info(msg string, keysAndValues ...interface{})  {}
func (l *TestLogger) Error(msg string, keysAndValues ...interface{}) {}
func (l *TestLogger) Debug(msg string, keysAndValues ...interface{}) {}

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
	Meta    operation.OperationMetadata
	Logger  operation.Logger
}

func (op *TestOperation) Execute(ctx context.Context) (string, error) {
	return operation.ExecuteOperation(op, func() (string, error) {
		return op.Service.DoSomething(ctx, op.Params)
	})
}

func (op *TestOperation) Metadata() operation.OperationMetadata {
	return op.Meta
}

func (op *TestOperation) Descriptor() operation.OperationDescriptor {
	return operation.OperationDescriptor{
		Type:     "TestOperation",
		Params:   op.Params,
		Metadata: op.Meta,
	}
}

func (op *TestOperation) GetMetadata() *operation.OperationMetadata { return &op.Meta }
func (op *TestOperation) GetLogger() operation.Logger               { return op.Logger }

func TestOperationFramework(t *testing.T) {
	// Setup
	registry := operation.NewServiceRegistry()
	operation.RegisterService[TestService](registry, &MockTestService{})

	logger := &TestLogger{}
	bus := operation.NewOperationBus(registry, logger)

	// Test operation creation
	op, err := operation.CreateOperation[*TestOperation](bus, "test input")
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
	registry := operation.NewServiceRegistry()
	
	// Test registration and retrieval
	mockService := &MockTestService{}
	operation.RegisterService[TestService](registry, mockService)
	
	retrieved := operation.GetService[TestService](registry)
	if retrieved != mockService {
		t.Error("Retrieved service should be the same instance as registered")
	}
}