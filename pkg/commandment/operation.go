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
	Params   interface{}       `json:"params"`
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

func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

// Logger defines the interface for structured logging used throughout the operation framework.
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

// ExecuteOperation is a generic execution wrapper that handles common logging and metadata
// for operations. This is typically used by concrete operation implementations.
func ExecuteOperation[T any](op OperationWithMetadata, businessLogic func() (T, error)) (T, error) {
	op.GetMetadata().Executed = time.Now()

	opTypeName := reflect.TypeOf(op).Elem().Name()
	logger := op.GetLogger()
	metadata := op.GetMetadata()

	logger.Info("Operation execution started",
		"operation_type", opTypeName,
		"operation_id", metadata.UUID,
	)

	result, err := businessLogic()
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
