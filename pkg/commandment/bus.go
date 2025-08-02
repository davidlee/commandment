package commandment

import (
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

// CreateOperation creates a new operation instance with injected dependencies.
// This is the core method that uses reflection to instantiate operations with
// their required services, metadata, and logger.
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

// DescriptorFactory recreates an executable operation from a serialized descriptor.
// This method must be implemented by users for their specific operation types.
type DescriptorFactory interface {
	CreateFromDescriptor(descriptor OperationDescriptor) (interface{}, error)
}

// getRequiredServiceType extracts the service type from an operation type using reflection.
func getRequiredServiceType[TOp any]() reflect.Type {
	opType := reflect.TypeOf((*TOp)(nil)).Elem()
	if opType.Kind() == reflect.Ptr {
		opType = opType.Elem()
	}
	// Convention: look for Service field
	serviceField, _ := opType.FieldByName("Service")
	return serviceField.Type
}

// newOperationWithService creates an operation instance using reflection.
func newOperationWithService[TOp any](params, service any, metadata OperationMetadata, logger Logger) (TOp, error) {
	opType := reflect.TypeOf((*TOp)(nil)).Elem()

	var opValue reflect.Value
	if opType.Kind() == reflect.Ptr {
		structType := opType.Elem()
		opValue = reflect.New(structType)
	} else {
		opValue = reflect.New(opType).Elem()
	}

	// Get the struct value to set fields on
	structValue := opValue
	if opType.Kind() == reflect.Ptr {
		structValue = opValue.Elem()
	}

	structValue.FieldByName("Params").Set(reflect.ValueOf(params))
	structValue.FieldByName("Service").Set(reflect.ValueOf(service))
	structValue.FieldByName("Meta").Set(reflect.ValueOf(metadata))
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
