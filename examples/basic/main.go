package main

import (
	"context"
	"fmt"
	"log"
	"os"

	charmlog "github.com/charmbracelet/log"

	"github.com/davidlee/commandment/examples/nodemanager"
	"github.com/davidlee/commandment/pkg/commandment"
)

// SimpleLogger provides a basic implementation of the Logger interface.
type SimpleLogger struct {
	logger *charmlog.Logger
}

func NewSimpleLogger() *SimpleLogger {
	logger := charmlog.New(os.Stderr)
	return &SimpleLogger{
		logger: logger.With("component", "operation-bus"),
	}
}

func (l *SimpleLogger) Info(msg string, keysAndValues ...any) {
	l.logger.Info(msg, keysAndValues...)
}

func (l *SimpleLogger) Warn(msg string, keysAndValues ...any) {
	l.logger.Warn(msg, keysAndValues...)
}

func (l *SimpleLogger) Error(msg string, keysAndValues ...any) {
	l.logger.Error(msg, keysAndValues...)
}

func (l *SimpleLogger) Debug(msg string, keysAndValues ...any) {
	l.logger.Debug(msg, keysAndValues...)
}

func main() {
	fmt.Println("🚀 Commandment POC - Operation Pattern Demo")
	fmt.Println("==========================================")

	// Setup the operation framework
	logger := NewSimpleLogger()
	registry := commandment.NewServiceRegistry()

	// Register domain services
	commandment.RegisterService[nodemanager.TreeService](registry, nodemanager.NewMockTreeService())
	commandment.RegisterService[nodemanager.ListService](registry, nodemanager.NewMockListService())
	commandment.RegisterService[nodemanager.NodeService](registry, nodemanager.NewMockNodeService())

	// Create operation bus and domain-specific wrapper
	operationBus := commandment.NewOperationBus(registry, logger)
	nodeManagerBus := nodemanager.NewNodeManagerBus(operationBus)

	fmt.Println("\n1. 🌳 Executing DisplayNodeTreeCommand...")
	executeTreeDisplay(nodeManagerBus)

	fmt.Println("\n2. 📝 Executing CreateListCommand...")
	executeListCreation(nodeManagerBus)

	fmt.Println("\n3. 👁️  Executing ShowNodeQuery...")
	executeShowNode(nodeManagerBus)

	fmt.Println("\n4. 🔒 Testing Query-Only Access...")
	testQueryOnlyAccess(nodeManagerBus)

	fmt.Println("\n✅ Demo completed!")
}

func executeTreeDisplay(invoker nodemanager.OperationInvoker) {
	params := nodemanager.DisplayNodeTreeCommandParams{
		RootReference: "root-123",
		MaxDepth:      3,
	}

	cmd, err := invoker.NewDisplayNodeTreeCommand(params)
	if err != nil {
		log.Fatalf("Failed to create command: %v", err)
	}

	fmt.Printf("   Command ID: %s\n", cmd.Metadata().UUID)

	result, err := cmd.Execute(context.Background())
	if err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}

	fmt.Printf("   📊 Result: %d nodes, max depth %d\n", len(result.Nodes), result.Stats.MaxDepth)
	for _, node := range result.Nodes {
		fmt.Printf("   - %s (ID: %d)\n", node.Title, node.ID)
	}
}

func executeListCreation(invoker nodemanager.OperationInvoker) {
	params := nodemanager.CreateListCommandParams{
		Title:       "My New List",
		Description: "A list created via command pattern",
		ParentID:    nil,
	}

	cmd, err := invoker.NewCreateListCommand(params)
	if err != nil {
		log.Fatalf("Failed to create command: %v", err)
	}

	result, err := cmd.Execute(context.Background())
	if err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}

	if len(result.Errors) > 0 {
		fmt.Printf("   ❌ Validation errors:\n")
		for _, errMsg := range result.Errors {
			fmt.Printf("   - %s: %s\n", errMsg.Field, errMsg.Message)
		}
	} else {
		fmt.Printf("   ✅ Created list: %s (ID: %d)\n", result.Node.Title, result.Node.ID)
	}
}

func executeShowNode(invoker nodemanager.OperationInvoker) {
	params := nodemanager.ShowNodeQueryParams{
		Ref: 42,
	}

	query, err := invoker.NewShowNodeQuery(params)
	if err != nil {
		log.Fatalf("Failed to create query: %v", err)
	}

	result, err := query.Execute(context.Background())
	if err != nil {
		log.Fatalf("Query execution failed: %v", err)
	}

	fmt.Printf("   📄 Node: %s (ID: %d)\n", result.Title, result.ID)
	fmt.Printf("   📝 Description: %s\n", result.Description)
}

func testQueryOnlyAccess(invoker nodemanager.OperationInvoker) {
	// Cast to query-only interface
	queryInvoker := nodemanager.QueryInvoker(invoker)

	params := nodemanager.ShowNodeQueryParams{Ref: 123}
	query, err := queryInvoker.NewShowNodeQuery(params)
	if err != nil {
		log.Fatalf("Failed to create query: %v", err)
	}

	result, err := query.Execute(context.Background())
	if err != nil {
		log.Fatalf("Query execution failed: %v", err)
	}

	fmt.Printf("   🔒 Query-only access worked: %s\n", result.Title)
	fmt.Printf("   💡 Query-only interface cannot access commands (compile-time safety)\n")
}
