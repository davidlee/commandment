// Package operation implements the Command/Query pattern with type-safe operation creation,
// service injection, and centralized logging. It provides a clean separation between
// clients, commands/queries, and business services with support for serialization
// and metadata tracking.
package operation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// OperationBus is the central orchestrator that manages service registry,
// creates operations with dependency injection, and handles operation lifecycle.
type OperationBus struct {
	registry *ServiceRegistry
	logger   Logger
}

// NewOperationBus creates a new OperationBus with the provided service registry and logger.
func NewOperationBus(registry *ServiceRegistry, logger Logger) *OperationBus {
	return &OperationBus{
		registry: registry,
		logger:   logger,
	}
}

// QueryInvoker provides methods for creating read-only query operations.
type QueryInvoker interface {
	NewShowNodeQuery(params ShowNodeQueryParams) (*ShowNodeQuery, error)
}

// CommandInvoker provides methods for creating command operations that mutate state.
type CommandInvoker interface {
	NewDisplayNodeTreeCommand(params DisplayNodeTreeCommandParams) (*DisplayNodeTreeCommand, error)
	NewCreateListCommand(params CreateListCommandParams) (*CreateListCommand, error)
}

// OperationInvoker combines QueryInvoker and CommandInvoker for full operation creation capabilities.
type OperationInvoker interface {
	QueryInvoker
	CommandInvoker
}

// OperationBus implements OperationInvoker by delegating to generic method
func (b *OperationBus) NewShowNodeQuery(params ShowNodeQueryParams) (*ShowNodeQuery, error) {
	return CreateOperation[*ShowNodeQuery](b, params)
}

func (b *OperationBus) NewDisplayNodeTreeCommand(params DisplayNodeTreeCommandParams) (*DisplayNodeTreeCommand, error) {
	return CreateOperation[*DisplayNodeTreeCommand](b, params)
}

func (b *OperationBus) NewCreateListCommand(params CreateListCommandParams) (*CreateListCommand, error) {
	return CreateOperation[*CreateListCommand](b, params)
}

// Generic method - used internally by OperationInvoker methods
func CreateOperation[TOp Operation[TResult], TResult any](
	bus *OperationBus,
	params any,
) (TOp, error) {
	// Use reflection to determine required service type
	serviceType := getRequiredServiceType[TOp]()
	service := bus.registry.get(serviceType)

	// Create metadata for new operation
	metadata := OperationMetadata{
		UUID:    generateUUID(),
		Created: time.Now(),
	}

	// Log operation creation
	opTypeName := reflect.TypeOf((*TOp)(nil)).Elem().Name()
	bus.logger.Info("Operation created",
		"operation_type", opTypeName,
		"operation_id", metadata.UUID,
		"service_type", serviceType.Name(),
	)

	// Create operation with injected service, metadata, and logger
	op, err := newOperationWithService[TOp](params, service, metadata, bus.logger)
	if err != nil {
		bus.logger.Error("Operation creation failed",
			"operation_type", opTypeName,
			"operation_id", metadata.UUID,
			"error", err,
		)
		return op, err
	}

	return op, nil
}

// Helper function to extract service type from operation type
func getRequiredServiceType[TOp any]() reflect.Type {
	opType := reflect.TypeOf((*TOp)(nil)).Elem()
	// If it's a pointer type, get the underlying struct
	if opType.Kind() == reflect.Ptr {
		opType = opType.Elem()
	}
	// Convention: look for Service field
	serviceField, _ := opType.FieldByName("Service")
	return serviceField.Type
}

// Factory function for operation creation
func newOperationWithService[TOp any](params, service any, metadata OperationMetadata, logger Logger) (TOp, error) {
	// Use reflection to create operation instance with params, service, metadata, and logger
	opType := reflect.TypeOf((*TOp)(nil)).Elem()

	var opValue reflect.Value
	if opType.Kind() == reflect.Ptr {
		// If TOp is *SomeStruct, create a new SomeStruct and get its pointer
		structType := opType.Elem()
		opValue = reflect.New(structType)
	} else {
		// If TOp is SomeStruct, create a new SomeStruct
		opValue = reflect.New(opType).Elem()
	}

	// Get the struct value to set fields on
	structValue := opValue
	if opType.Kind() == reflect.Ptr {
		structValue = opValue.Elem()
	}

	structValue.FieldByName("Params").Set(reflect.ValueOf(params))
	structValue.FieldByName("Service").Set(reflect.ValueOf(service))
	structValue.FieldByName("OperationMeta").Set(reflect.ValueOf(metadata))
	structValue.FieldByName("Logger").Set(reflect.ValueOf(logger))

	if opType.Kind() == reflect.Ptr {
		result, ok := opValue.Interface().(TOp)
		if !ok {
			var zero TOp
			return zero, fmt.Errorf("type assertion failed: got %T, expected %T", opValue.Interface(), zero)
		}
		return result, nil
	} else {
		result, ok := opValue.Addr().Interface().(TOp)
		if !ok {
			var zero TOp
			return zero, fmt.Errorf("type assertion failed: got %T, expected %T", opValue.Addr().Interface(), zero)
		}
		return result, nil
	}
}

// Deserialization: recreate executable operations from descriptors
func (b *OperationBus) CreateFromDescriptor(descriptor OperationDescriptor) (interface{}, error) {
	// Map command type to creation function
	switch descriptor.Type {
	case "DisplayNodeTreeCommand":
		params := DisplayNodeTreeCommandParams{}
		if err := json.Unmarshal(mustMarshal(descriptor.Params), &params); err != nil {
			return nil, err
		}
		return b.createDisplayNodeTreeCommand(params, descriptor.Metadata), nil

	case "CreateListCommand":
		params := CreateListCommandParams{}
		if err := json.Unmarshal(mustMarshal(descriptor.Params), &params); err != nil {
			return nil, err
		}
		return b.createCreateListCommand(params, descriptor.Metadata), nil

	case "ShowNodeQuery":
		params := ShowNodeQueryParams{}
		if err := json.Unmarshal(mustMarshal(descriptor.Params), &params); err != nil {
			return nil, err
		}
		return b.createShowNodeQuery(params, descriptor.Metadata), nil

	default:
		return nil, fmt.Errorf("unknown operation type: %s", descriptor.Type)
	}
}

func (b *OperationBus) createDisplayNodeTreeCommand(params DisplayNodeTreeCommandParams, metadata OperationMetadata) *DisplayNodeTreeCommand {
	service := GetService[TreeService](b.registry)
	return &DisplayNodeTreeCommand{
		Params:        params,
		Service:       service,
		OperationMeta: metadata,
		Logger:        b.logger,
	}
}

func (b *OperationBus) createCreateListCommand(params CreateListCommandParams, metadata OperationMetadata) *CreateListCommand {
	service := GetService[ListService](b.registry)
	return &CreateListCommand{
		Params:        params,
		Service:       service,
		OperationMeta: metadata,
		Logger:        b.logger,
	}
}

func (b *OperationBus) createShowNodeQuery(params ShowNodeQueryParams, metadata OperationMetadata) *ShowNodeQuery {
	service := GetService[NodeService](b.registry)
	return &ShowNodeQuery{
		Params:        params,
		Service:       service,
		OperationMeta: metadata,
		Logger:        b.logger,
	}
}
