package commandment

import (
	"fmt"
	"reflect"
	"time"
)

// OperationBus is the central orchestrator that manages service registry,
// creates operations with dependency injection, and handles operation lifecycle.
type OperationBus struct {
	registry    *ServiceRegistry
	logger      Logger
	defaultDeps any // Optional default Dependencies for all operations
}

// NewOperationBus creates a new OperationBus with the provided service registry and logger.
func NewOperationBus(registry *ServiceRegistry, logger Logger) *OperationBus {
	return &OperationBus{
		registry:    registry,
		logger:      logger,
		defaultDeps: nil,
	}
}

// NewOperationBusWithDefaultDependencies creates a new OperationBus with default Dependencies
// that will be available to all operations created by this bus.
func NewOperationBusWithDefaultDependencies(registry *ServiceRegistry, logger Logger, defaultDeps any) *OperationBus {
	return &OperationBus{
		registry:    registry,
		logger:      logger,
		defaultDeps: defaultDeps,
	}
}

// CreateOperation creates a new operation instance with injected dependencies.
// This is the core method that uses reflection to instantiate operations with
// their required services, metadata, and logger.
func CreateOperation[TOp Operation[TResult], TResult any](
	bus *OperationBus,
	params any,
) (TOp, error) {
	return createOperationInternal[TOp, TResult](bus, params, bus.defaultDeps)
}

// CreateOperationWithDependencies creates a new operation instance with specific Dependencies,
// overriding any default Dependencies configured on the bus.
func CreateOperationWithDependencies[TOp Operation[TResult], TResult any](
	bus *OperationBus,
	params any,
	deps any,
) (TOp, error) {
	return createOperationInternal[TOp, TResult](bus, params, deps)
}

// createOperationInternal is the shared implementation for operation creation
func createOperationInternal[TOp Operation[TResult], TResult any](
	bus *OperationBus,
	params any,
	deps any,
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
	logData := []any{
		"operation_type", opTypeName,
		"operation_id", metadata.UUID,
		"service_type", serviceType.Name(),
	}
	if deps != nil {
		depsType := reflect.TypeOf(deps).String()
		logData = append(logData, "dependencies_type", depsType)
	}
	bus.logger.Info("Operation created", logData...)

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

	// Store dependencies in operation for later context enrichment
	if deps != nil {
		storeOperationDependencies(op, deps)
	}

	return op, nil
}

// DescriptorFactory recreates an executable operation from a serialized descriptor.
// This method must be implemented by users for their specific operation types.
type DescriptorFactory interface {
	CreateFromDescriptor(descriptor OperationDescriptor) (any, error)
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

// operationDependencies stores dependencies for operations using a weak map pattern
var operationDependencies = make(map[any]any)

// storeOperationDependencies associates dependencies with an operation instance
func storeOperationDependencies(op, deps any) {
	operationDependencies[op] = deps
}

// GetOperationDependencies retrieves dependencies for an operation instance
func GetOperationDependencies(op any) any {
	return operationDependencies[op]
}
