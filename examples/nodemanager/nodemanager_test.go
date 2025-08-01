package nodemanager_test

import (
	"context"
	"testing"

	"github.com/davidlee/commandment/examples/nodemanager"
	"github.com/davidlee/commandment/pkg/operation"
)

// Simple test logger
type TestLogger struct{}

func (l *TestLogger) Info(msg string, keysAndValues ...interface{})  {}
func (l *TestLogger) Error(msg string, keysAndValues ...interface{}) {}
func (l *TestLogger) Debug(msg string, keysAndValues ...interface{}) {}

func TestNodeManagerBasicFlow(t *testing.T) {
	// Setup framework
	registry := operation.NewServiceRegistry()
	operation.RegisterService[nodemanager.TreeService](registry, nodemanager.NewMockTreeService())
	operation.RegisterService[nodemanager.ListService](registry, nodemanager.NewMockListService())
	operation.RegisterService[nodemanager.NodeService](registry, nodemanager.NewMockNodeService())

	logger := &TestLogger{}
	operationBus := operation.NewOperationBus(registry, logger)
	
	// Create domain-specific bus
	nodeManagerBus := nodemanager.NewNodeManagerBus(operationBus)

	// Test command creation and execution
	params := nodemanager.DisplayNodeTreeCommandParams{
		RootReference: "test-root",
		MaxDepth:      2,
	}

	cmd, err := nodeManagerBus.NewDisplayNodeTreeCommand(params)
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
	operation.RegisterService[nodemanager.NodeService](registry, nodemanager.NewMockNodeService())

	logger := &TestLogger{}
	operationBus := operation.NewOperationBus(registry, logger)
	nodeManagerBus := nodemanager.NewNodeManagerBus(operationBus)

	// Cast to query-only interface
	queryInvoker := nodemanager.QueryInvoker(nodeManagerBus)

	// Test query creation and execution
	params := nodemanager.ShowNodeQueryParams{Ref: 42}
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