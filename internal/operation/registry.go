package operation

import (
	"fmt"
	"reflect"
)

type ServiceRegistry struct {
	services map[reflect.Type]interface{}
}

// NewServiceRegistry creates a new empty service registry.
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[reflect.Type]interface{}),
	}
}

// Non-generic helper functions for the registry
func (r *ServiceRegistry) register(serviceType reflect.Type, service interface{}) {
	r.services[serviceType] = service
}

func (r *ServiceRegistry) get(serviceType reflect.Type) interface{} {
	service, exists := r.services[serviceType]
	if !exists {
		panic(fmt.Sprintf("Service type %v not registered", serviceType))
	}
	return service
}

// RegisterService registers a service instance of type T in the registry.
func RegisterService[T any](r *ServiceRegistry, service T) {
	r.register(reflect.TypeOf((*T)(nil)).Elem(), service)
}

// GetService retrieves a service of type T from the registry.
func GetService[T any](r *ServiceRegistry) T {
	serviceType := reflect.TypeOf((*T)(nil)).Elem()
	service := r.get(serviceType)
	result, ok := service.(T)
	if !ok {
		var zero T
		panic(fmt.Sprintf("service type assertion failed: got %T, expected %T", service, zero))
	}
	return result
}
