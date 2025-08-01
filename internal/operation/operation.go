package operation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"time"
)

// Shared base interface for commands and queries
type Operation[TResult any] interface {
	Execute(ctx context.Context) (TResult, error)
	Metadata() OperationMetadata
	Descriptor() OperationDescriptor
}

// Commands mutate state
type Command[TResult any] interface {
	Operation[TResult]
}

// Queries are read-only
type Query[TResult any] interface {
	Operation[TResult]
}

// Optional metadata for audit trails and debugging
type OperationMetadata struct {
	UUID     string    `json:"uuid"`
	Created  time.Time `json:"created"`
	Executed time.Time `json:"executed,omitempty"`
	Returned time.Time `json:"returned,omitempty"`
}

// Serializable operation representation
type OperationDescriptor struct {
	Type     string            `json:"type"`
	Params   interface{}       `json:"params"`
	Metadata OperationMetadata `json:"metadata"`
}

// JSON serialization helper for type-safe params
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

// Logger interface
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

// Generic execution wrapper that handles common logging and metadata
func executeOperation[T any](op operationWithMetadata, businessLogic func() (T, error)) (T, error) {
	op.getMetadata().Executed = time.Now()

	opTypeName := reflect.TypeOf(op).Elem().Name()
	logger := op.getLogger()
	metadata := op.getMetadata()

	logger.Info("Operation execution started",
		"operation_type", opTypeName,
		"operation_id", metadata.UUID,
	)

	result, err := businessLogic()
	op.getMetadata().Returned = time.Now()

	duration := op.getMetadata().Returned.Sub(op.getMetadata().Executed)
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

// Helper interface for accessing operation metadata and logger
type operationWithMetadata interface {
	getMetadata() *OperationMetadata
	getLogger() Logger
}

func generateUUID() string {
	// Simple UUID alternative using crypto/rand
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
