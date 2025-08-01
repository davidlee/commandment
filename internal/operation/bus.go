package operation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type OperationBus struct {
	registry *ServiceRegistry
	logger   Logger
}

func NewOperationBus(registry *ServiceRegistry, logger Logger) *OperationBus {
	return &OperationBus{
		registry: registry,
		logger:   logger,
	}
}

// Query-only interface for read-only operations
type QueryInvoker interface {
	NewShowNodeQuery(params ShowNodeQueryParams) (*ShowNodeQuery, error)
}

// Command-only interface for mutations
type CommandInvoker interface {
	NewDisplayNodeTreeCommand(params DisplayNodeTreeCommandParams) (*DisplayNodeTreeCommand, error)
	NewCreateListCommand(params CreateListCommandParams) (*CreateListCommand, error)
}

// Full interface includes both queries and commands
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
func newOperationWithService[TOp any](params any, service any, metadata OperationMetadata, logger Logger) (TOp, error) {
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
		return opValue.Interface().(TOp), nil
	} else {
		return opValue.Addr().Interface().(TOp), nil
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
		return b.createDisplayNodeTreeCommand(params, descriptor.Metadata)

	case "CreateListCommand":
		params := CreateListCommandParams{}
		if err := json.Unmarshal(mustMarshal(descriptor.Params), &params); err != nil {
			return nil, err
		}
		return b.createCreateListCommand(params, descriptor.Metadata)

	case "ShowNodeQuery":
		params := ShowNodeQueryParams{}
		if err := json.Unmarshal(mustMarshal(descriptor.Params), &params); err != nil {
			return nil, err
		}
		return b.createShowNodeQuery(params, descriptor.Metadata)

	default:
		return nil, fmt.Errorf("unknown operation type: %s", descriptor.Type)
	}
}

func (b *OperationBus) createDisplayNodeTreeCommand(params DisplayNodeTreeCommandParams, metadata OperationMetadata) (*DisplayNodeTreeCommand, error) {
	service := GetService[TreeService](b.registry)
	return &DisplayNodeTreeCommand{
		Params:        params,
		Service:       service,
		OperationMeta: metadata,
		Logger:        b.logger,
	}, nil
}

func (b *OperationBus) createCreateListCommand(params CreateListCommandParams, metadata OperationMetadata) (*CreateListCommand, error) {
	service := GetService[ListService](b.registry)
	return &CreateListCommand{
		Params:        params,
		Service:       service,
		OperationMeta: metadata,
		Logger:        b.logger,
	}, nil
}

func (b *OperationBus) createShowNodeQuery(params ShowNodeQueryParams, metadata OperationMetadata) (*ShowNodeQuery, error) {
	service := GetService[NodeService](b.registry)
	return &ShowNodeQuery{
		Params:        params,
		Service:       service,
		OperationMeta: metadata,
		Logger:        b.logger,
	}, nil
}
