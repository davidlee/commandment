package commandment_test

import (
	"context"
	"testing"

	"github.com/davidlee/commandment/pkg/commandment"
)

// Test Dependencies type
type TestDependencies struct {
	Value string
	DB    *MockDB
}

func (d *TestDependencies) GetValue() string {
	return d.Value
}

// Mock database for testing
type MockDB struct {
	Connected bool
}

func (db *MockDB) Connect() {
	db.Connected = true
}

// Special Dependencies type for testing override
type SpecialDependencies struct {
	SpecialValue string
}

// Test service that uses Dependencies
type DependencyAwareService struct {
	name string
}

func (s *DependencyAwareService) ProcessWithDependencies(ctx context.Context, input string) (string, error) {
	deps := commandment.DependenciesFromContext(ctx)
	if deps == nil {
		return "no-deps:" + input, nil
	}
	
	testDeps, ok := deps.(*TestDependencies)
	if !ok {
		return "wrong-type:" + input, nil
	}
	
	return testDeps.GetValue() + ":" + input, nil
}

// Test operation that uses Dependencies from context
type DependencyAwareOperation struct {
	Params  string
	Service DependencyAwareService
	Meta    commandment.OperationMetadata
	Logger  commandment.Logger
}

func (op *DependencyAwareOperation) Execute(ctx context.Context) (string, error) {
	return commandment.ExecuteOperation(ctx, op, func(ctx context.Context) (string, error) {
		return op.Service.ProcessWithDependencies(ctx, op.Params)
	})
}

func (op *DependencyAwareOperation) Metadata() commandment.OperationMetadata {
	return op.Meta
}

func (op *DependencyAwareOperation) Descriptor() commandment.OperationDescriptor {
	return commandment.OperationDescriptor{
		Type:     "DependencyAwareOperation",
		Params:   op.Params,
		Metadata: op.Meta,
	}
}

func (op *DependencyAwareOperation) GetMetadata() *commandment.OperationMetadata { return &op.Meta }
func (op *DependencyAwareOperation) GetLogger() commandment.Logger               { return op.Logger }

// Test operation that accesses Dependencies directly via GetDependencies
type DirectDependencyOperation struct {
	Params  string
	Service DependencyAwareService
	Meta    commandment.OperationMetadata
	Logger  commandment.Logger
}

func (op *DirectDependencyOperation) Execute(ctx context.Context) (string, error) {
	return commandment.ExecuteOperation(ctx, op, func(ctx context.Context) (string, error) {
		// Access Dependencies directly from operation
		deps := commandment.GetDependencies(op)
		if deps == nil {
			return "no-deps-direct:" + op.Params, nil
		}
		
		testDeps, ok := deps.(*TestDependencies)
		if !ok {
			return "wrong-type-direct:" + op.Params, nil
		}
		
		return "direct:" + testDeps.GetValue() + ":" + op.Params, nil
	})
}

func (op *DirectDependencyOperation) Metadata() commandment.OperationMetadata {
	return op.Meta
}

func (op *DirectDependencyOperation) Descriptor() commandment.OperationDescriptor {
	return commandment.OperationDescriptor{
		Type:     "DirectDependencyOperation",
		Params:   op.Params,
		Metadata: op.Meta,
	}
}

func (op *DirectDependencyOperation) GetMetadata() *commandment.OperationMetadata { return &op.Meta }
func (op *DirectDependencyOperation) GetLogger() commandment.Logger               { return op.Logger }

func TestDependenciesWithDefaultDependencies(t *testing.T) {
	// Setup Dependencies
	deps := &TestDependencies{
		Value: "test-value",
		DB:    &MockDB{},
	}
	deps.DB.Connect()

	// Setup registry and bus with default Dependencies
	registry := commandment.NewServiceRegistry()
	commandment.RegisterService[DependencyAwareService](registry, DependencyAwareService{name: "test"})

	logger := &TestLogger{}
	bus := commandment.NewOperationBusWithDefaultDependencies(registry, logger, deps)

	// Test operation creation
	op, err := commandment.CreateOperation[*DependencyAwareOperation](bus, "test-input")
	if err != nil {
		t.Fatalf("Failed to create operation: %v", err)
	}

	if op == nil {
		t.Fatal("Operation should not be nil")
	}

	// Test execution with Dependencies access from context
	result, err := op.Execute(context.Background())
	if err != nil {
		t.Fatalf("Operation execution failed: %v", err)
	}

	expected := "test-value:test-input"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Verify Dependencies are accessible
	if !deps.DB.Connected {
		t.Error("Dependencies should be properly initialized")
	}
}

func TestDependenciesWithOverride(t *testing.T) {
	// Setup default Dependencies
	defaultDeps := &TestDependencies{
		Value: "default-value",
		DB:    &MockDB{},
	}

	// Setup special Dependencies for override
	specialDeps := &TestDependencies{
		Value: "special-value",
		DB:    &MockDB{},
	}
	specialDeps.DB.Connect()

	// Setup registry and bus with default Dependencies
	registry := commandment.NewServiceRegistry()
	commandment.RegisterService[DependencyAwareService](registry, DependencyAwareService{name: "test"})

	logger := &TestLogger{}
	bus := commandment.NewOperationBusWithDefaultDependencies(registry, logger, defaultDeps)

	// Test operation creation with Dependencies override
	op, err := commandment.CreateOperationWithDependencies[*DependencyAwareOperation](bus, "override-input", specialDeps)
	if err != nil {
		t.Fatalf("Failed to create operation with Dependencies override: %v", err)
	}

	// Test execution - should use special Dependencies, not default
	result, err := op.Execute(context.Background())
	if err != nil {
		t.Fatalf("Operation execution failed: %v", err)
	}

	expected := "special-value:override-input"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestDependenciesDirectAccess(t *testing.T) {
	// Setup Dependencies
	deps := &TestDependencies{
		Value: "direct-test",
		DB:    &MockDB{},
	}

	// Setup registry and bus
	registry := commandment.NewServiceRegistry()
	commandment.RegisterService[DependencyAwareService](registry, DependencyAwareService{name: "test"})

	logger := &TestLogger{}
	bus := commandment.NewOperationBusWithDefaultDependencies(registry, logger, deps)

	// Test operation creation
	op, err := commandment.CreateOperation[*DirectDependencyOperation](bus, "direct-input")
	if err != nil {
		t.Fatalf("Failed to create operation: %v", err)
	}

	// Test execution with direct Dependencies access
	result, err := op.Execute(context.Background())
	if err != nil {
		t.Fatalf("Operation execution failed: %v", err)
	}

	expected := "direct:direct-test:direct-input"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestNoDependencies(t *testing.T) {
	// Setup registry and bus without Dependencies
	registry := commandment.NewServiceRegistry()
	commandment.RegisterService[DependencyAwareService](registry, DependencyAwareService{name: "test"})

	logger := &TestLogger{}
	bus := commandment.NewOperationBus(registry, logger) // No Dependencies

	// Test operation creation
	op, err := commandment.CreateOperation[*DependencyAwareOperation](bus, "no-deps-input")
	if err != nil {
		t.Fatalf("Failed to create operation: %v", err)
	}

	// Test execution without Dependencies
	result, err := op.Execute(context.Background())
	if err != nil {
		t.Fatalf("Operation execution failed: %v", err)
	}

	expected := "no-deps:no-deps-input"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestDependenciesFromContext(t *testing.T) {
	deps := &TestDependencies{Value: "context-test"}
	
	// Test context enrichment
	ctx := context.Background()
	enrichedCtx := commandment.WithDependencies(ctx, deps)
	
	// Test Dependencies retrieval
	retrieved := commandment.DependenciesFromContext(enrichedCtx)
	if retrieved == nil {
		t.Fatal("Dependencies should be retrievable from context")
	}
	
	testDeps, ok := retrieved.(*TestDependencies)
	if !ok {
		t.Fatalf("Expected *TestDependencies, got %T", retrieved)
	}
	
	if testDeps.Value != "context-test" {
		t.Errorf("Expected %q, got %q", "context-test", testDeps.Value)
	}
}

func TestDependenciesFromEmptyContext(t *testing.T) {
	ctx := context.Background()
	
	// Test Dependencies retrieval from empty context
	deps := commandment.DependenciesFromContext(ctx)
	if deps != nil {
		t.Errorf("Expected nil Dependencies from empty context, got %T", deps)
	}
}

func TestGetDependenciesFromOperation(t *testing.T) {
	// Setup Dependencies
	deps := &TestDependencies{Value: "get-test"}

	// Setup registry and bus
	registry := commandment.NewServiceRegistry()
	commandment.RegisterService[DependencyAwareService](registry, DependencyAwareService{name: "test"})

	logger := &TestLogger{}
	bus := commandment.NewOperationBusWithDefaultDependencies(registry, logger, deps)

	// Create operation
	op, err := commandment.CreateOperation[*DependencyAwareOperation](bus, "get-input")
	if err != nil {
		t.Fatalf("Failed to create operation: %v", err)
	}

	// Test direct Dependencies access from operation
	retrieved := commandment.GetDependencies(op)
	if retrieved == nil {
		t.Fatal("Dependencies should be retrievable from operation")
	}

	testDeps, ok := retrieved.(*TestDependencies)
	if !ok {
		t.Fatalf("Expected *TestDependencies, got %T", retrieved)
	}

	if testDeps.Value != "get-test" {
		t.Errorf("Expected %q, got %q", "get-test", testDeps.Value)
	}
}

func TestMultipleDependencyTypes(t *testing.T) {
	// Setup different Dependencies types
	testDeps := &TestDependencies{Value: "test-type"}
	specialDeps := &SpecialDependencies{SpecialValue: "special-type"}

	// Setup registry and bus with test Dependencies as default
	registry := commandment.NewServiceRegistry()
	commandment.RegisterService[DependencyAwareService](registry, DependencyAwareService{name: "test"})

	logger := &TestLogger{}
	bus := commandment.NewOperationBusWithDefaultDependencies(registry, logger, testDeps)

	// Create operation with default Dependencies
	op1, err := commandment.CreateOperation[*DependencyAwareOperation](bus, "default")
	if err != nil {
		t.Fatalf("Failed to create operation: %v", err)
	}

	// Create operation with special Dependencies override
	op2, err := commandment.CreateOperationWithDependencies[*DependencyAwareOperation](bus, "special", specialDeps)
	if err != nil {
		t.Fatalf("Failed to create operation with override: %v", err)
	}

	// Test that each operation has the correct Dependencies type
	deps1 := commandment.GetDependencies(op1)
	if _, ok := deps1.(*TestDependencies); !ok {
		t.Errorf("Operation 1 should have TestDependencies, got %T", deps1)
	}

	deps2 := commandment.GetDependencies(op2)
	if _, ok := deps2.(*SpecialDependencies); !ok {
		t.Errorf("Operation 2 should have SpecialDependencies, got %T", deps2)
	}
}