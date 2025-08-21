// Package commandment implements the Command/Query pattern with type-safe operation creation,
// service injection, and centralized logging. It provides a clean separation between
// clients, commands/queries, and business services with support for serialization
// and metadata tracking.
package commandment

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"time"
)

// contextKey is a private type for context keys to avoid collisions
type contextKey string

// operationMetadataKey is the context key for operation metadata
const operationMetadataKey contextKey = "commandment:operation:metadata"

// dependenciesKey is the context key for dependencies
const dependenciesKey contextKey = "commandment:dependencies"

// Operation is the shared base interface for commands and queries,
// providing common behavior for execution, metadata access, and serialization.
type Operation[TResult any] interface {
	Execute(ctx context.Context) (TResult, error)
	Metadata() OperationMetadata
	Descriptor() OperationDescriptor
}

// Command extends Operation for operations that mutate state.
type Command[TResult any] interface {
	Operation[TResult]
}

// Query extends Operation for read-only operations that don't mutate state.
type Query[TResult any] interface {
	Operation[TResult]
}

// OperationMetadata contains timestamps and identifiers for audit trails and debugging.
type OperationMetadata struct {
	UUID     string    `json:"uuid"`
	Created  time.Time `json:"created"`
	Executed time.Time `json:"executed,omitempty"`
	Returned time.Time `json:"returned,omitempty"`
}

// OperationDescriptor provides a serializable representation of an operation
// including its type, parameters, and metadata for persistence and reconstruction.
type OperationDescriptor struct {
	Type     string            `json:"type"`
	Params   any       `json:"params"`
	Metadata OperationMetadata `json:"metadata"`
}

// MarshalJSON provides custom JSON serialization for type-safe parameter marshaling.
func (od OperationDescriptor) MarshalJSON() ([]byte, error) {
	type Alias OperationDescriptor
	return json.Marshal(&struct {
		*Alias
		Params json.RawMessage `json:"params"`
	}{
		Alias:  (*Alias)(&od),
		Params: mustMarshal(od.Params),
	})
}

func mustMarshal(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

// Logger defines the interface for structured logging used throughout the operation framework.
type Logger interface {
	Info(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Debug(msg string, keysAndValues ...any)
}

// WithOperationMetadata adds operation metadata to the context
func WithOperationMetadata(ctx context.Context, meta *OperationMetadata) context.Context {
	return context.WithValue(ctx, operationMetadataKey, meta)
}

// OperationMetadataFromContext retrieves operation metadata from context
func OperationMetadataFromContext(ctx context.Context) *OperationMetadata {
	if meta, ok := ctx.Value(operationMetadataKey).(*OperationMetadata); ok {
		return meta
	}
	return nil
}

// WithDependencies adds dependencies to the context
func WithDependencies(ctx context.Context, deps any) context.Context {
	return context.WithValue(ctx, dependenciesKey, deps)
}

// DependenciesFromContext retrieves dependencies from context.
// Returns nil if no dependencies are available.
func DependenciesFromContext(ctx context.Context) any {
	return ctx.Value(dependenciesKey)
}

// GetDependencies retrieves dependencies from an operation instance.
// This is a convenience function for accessing dependencies outside of execution context.
// During execution, prefer DependenciesFromContext(ctx) for context-based access.
func GetDependencies(op OperationWithMetadata) any {
	return GetOperationDependencies(op)
}

// ExecuteOperation is a context-aware execution wrapper that enriches context with operation metadata
// before calling the business logic. This allows downstream services to access operation metadata.
func ExecuteOperation[T any](ctx context.Context, op OperationWithMetadata, businessLogic func(context.Context) (T, error)) (T, error) {
	op.GetMetadata().Executed = time.Now()

	opTypeName := reflect.TypeOf(op).Elem().Name()
	logger := op.GetLogger()
	metadata := op.GetMetadata()

	// Enrich context with operation metadata
	ctxWithMeta := WithOperationMetadata(ctx, metadata)
	
	// Enrich context with dependencies if available
	deps := GetOperationDependencies(op)
	if deps != nil {
		ctxWithMeta = WithDependencies(ctxWithMeta, deps)
	}

	logger.Info("Operation execution started",
		"operation_type", opTypeName,
		"operation_id", metadata.UUID,
	)

	result, err := businessLogic(ctxWithMeta)
	op.GetMetadata().Returned = time.Now()

	duration := op.GetMetadata().Returned.Sub(op.GetMetadata().Executed)
	if err != nil {
		logger.Error("Operation execution failed",
			"operation_type", opTypeName,
			"operation_id", metadata.UUID,
			"duration_ms", duration.Milliseconds(),
			"error", err,
		)
	} else {
		logger.Info("Operation execution completed",
			"operation_type", opTypeName,
			"operation_id", metadata.UUID,
			"duration_ms", duration.Milliseconds(),
		)
	}

	return result, err
}

// OperationWithMetadata is a helper interface for accessing operation metadata and logger.
// Concrete operations should implement this interface to work with ExecuteOperation.
type OperationWithMetadata interface {
	GetMetadata() *OperationMetadata
	GetLogger() Logger
}

func generateUUID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("failed to generate random bytes for UUID: " + err.Error())
	}
	return hex.EncodeToString(bytes)
}
