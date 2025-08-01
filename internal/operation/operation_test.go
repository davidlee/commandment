package operation_test

import (
	"context"
	"testing"

	"com.github/davidlee/commandment/internal/operation"
	"com.github/davidlee/commandment/internal/services"
)

// Simple test logger
type TestLogger struct{}

func (l *TestLogger) Info(msg string, keysAndValues ...interface{})  {}
func (l *TestLogger) Error(msg string, keysAndValues ...interface{}) {}
func (l *TestLogger) Debug(msg string, keysAndValues ...interface{}) {}

func TestOperationBusBasicFlow(t *testing.T) {
	// Setup
	registry := operation.NewServiceRegistry()
	operation.RegisterService[operation.TreeService](registry, services.NewMockTreeService())
	operation.RegisterService[operation.ListService](registry, services.NewMockListService())
	operation.RegisterService[operation.NodeService](registry, services.NewMockNodeService())

	logger := &TestLogger{}
	bus := operation.NewOperationBus(registry, logger)

	// Test command creation and execution
	params := operation.DisplayNodeTreeCommandParams{
		RootReference: "test-root",
		MaxDepth:      2,
	}

	cmd, err := bus.NewDisplayNodeTreeCommand(params)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	if cmd == nil {
		t.Fatal("Command should not be nil")
	}

	// Test execution
	result, err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected some nodes in result")
	}

	// Test metadata
	metadata := cmd.Metadata()
	if metadata.UUID == "" {
		t.Error("Expected UUID to be set")
	}

	if metadata.Created.IsZero() {
		t.Error("Expected Created timestamp to be set")
	}
}

func TestQueryOnlyInterface(t *testing.T) {
	// Setup
	registry := operation.NewServiceRegistry()
	operation.RegisterService[operation.NodeService](registry, services.NewMockNodeService())

	logger := &TestLogger{}
	bus := operation.NewOperationBus(registry, logger)

	// Cast to query-only interface
	queryInvoker := operation.QueryInvoker(bus)

	// Test query creation and execution
	params := operation.ShowNodeQueryParams{Ref: 42}
	query, err := queryInvoker.NewShowNodeQuery(params)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	result, err := query.Execute(context.Background())
	if err != nil {
		t.Fatalf("Query execution failed: %v", err)
	}

	if result.ID != 42 {
		t.Errorf("Expected node ID 42, got %d", result.ID)
	}
}

func TestSerialization(t *testing.T) {
	// Setup
	registry := operation.NewServiceRegistry()
	operation.RegisterService[operation.ListService](registry, services.NewMockListService())

	logger := &TestLogger{}
	bus := operation.NewOperationBus(registry, logger)

	// Create command
	params := operation.CreateListCommandParams{
		Title:       "Test List",
		Description: "A test list",
	}

	cmd, err := bus.NewCreateListCommand(params)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Get descriptor
	descriptor := cmd.Descriptor()
	if descriptor.Type != "CreateListCommand" {
		t.Errorf("Expected type CreateListCommand, got %s", descriptor.Type)
	}

	// Test deserialization
	recreated, err := bus.CreateFromDescriptor(descriptor)
	if err != nil {
		t.Fatalf("Failed to recreate command: %v", err)
	}

	recreatedCmd, ok := recreated.(*operation.CreateListCommand)
	if !ok {
		t.Fatalf("Expected *CreateListCommand, got %T", recreated)
	}

	// Execute recreated command
	result, err := recreatedCmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("Recreated command execution failed: %v", err)
	}

	if result.Node.Title != "Test List" {
		t.Errorf("Expected title 'Test List', got %s", result.Node.Title)
	}
}
