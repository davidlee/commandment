package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	charmlog "github.com/charmbracelet/log"

	"com.github/davidlee/commandment/internal/operation"
	"com.github/davidlee/commandment/internal/services"
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

func (l *SimpleLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, keysAndValues...)
}

func (l *SimpleLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Error(msg, keysAndValues...)
}

func (l *SimpleLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debug(msg, keysAndValues...)
}

func main() {
	fmt.Println("🚀 Commandment POC - Operation Pattern Demo")
	fmt.Println("==========================================")

	// Setup
	logger := NewSimpleLogger()
	registry := operation.NewServiceRegistry()

	// Register mock services
	operation.RegisterService[operation.TreeService](registry, services.NewMockTreeService())
	operation.RegisterService[operation.ListService](registry, services.NewMockListService())
	operation.RegisterService[operation.NodeService](registry, services.NewMockNodeService())

	// Create operation bus
	bus := operation.NewOperationBus(registry, logger)

	fmt.Println("\n1. 🌳 Executing DisplayNodeTreeCommand...")
	executeTreeDisplay(bus)

	fmt.Println("\n2. 📝 Executing CreateListCommand...")
	executeListCreation(bus)

	fmt.Println("\n3. 👁️  Executing ShowNodeQuery...")
	executeShowNode(bus)

	fmt.Println("\n4. 📦 Testing Serialization...")
	testSerialization(bus)

	fmt.Println("\n5. 🔒 Testing Query-Only Access...")
	testQueryOnlyAccess(bus)

	fmt.Println("\n✅ Demo completed!")
}

func executeTreeDisplay(invoker operation.OperationInvoker) {
	params := operation.DisplayNodeTreeCommandParams{
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

func executeListCreation(invoker operation.OperationInvoker) {
	params := operation.CreateListCommandParams{
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

func executeShowNode(invoker operation.OperationInvoker) {
	params := operation.ShowNodeQueryParams{
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

func testSerialization(bus *operation.OperationBus) {
	// Create a command
	params := operation.CreateListCommandParams{
		Title:       "Serialization Test",
		Description: "Testing command serialization",
		ParentID:    nil,
	}

	cmd, err := bus.NewCreateListCommand(params)
	if err != nil {
		log.Fatalf("Failed to create command: %v", err)
	}

	// Serialize to JSON
	descriptor := cmd.Descriptor()
	jsonData, err := json.MarshalIndent(descriptor, "   ", "  ")
	if err != nil {
		log.Fatalf("Failed to serialize: %v", err)
	}

	fmt.Printf("   📤 Serialized command:\n%s\n", string(jsonData))

	// Deserialize and execute
	recreatedCmd, err := bus.CreateFromDescriptor(descriptor)
	if err != nil {
		log.Fatalf("Failed to deserialize: %v", err)
	}

	switch c := recreatedCmd.(type) {
	case *operation.CreateListCommand:
		result, err := c.Execute(context.Background())
		if err != nil {
			log.Fatalf("Recreated command execution failed: %v", err)
		}
		fmt.Printf("   📥 Deserialized command executed successfully: %s\n", result.Node.Title)
	default:
		log.Fatalf("Unexpected command type: %T", recreatedCmd)
	}
}

func testQueryOnlyAccess(invoker operation.OperationInvoker) {
	// Cast to query-only interface
	queryInvoker := operation.QueryInvoker(invoker)

	params := operation.ShowNodeQueryParams{Ref: 123}
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
