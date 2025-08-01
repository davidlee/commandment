package operation

import (
	"fmt"
	"reflect"
)

type ServiceRegistry struct {
	services map[reflect.Type]interface{}
}

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

// Generic wrapper functions
func RegisterService[T any](r *ServiceRegistry, service T) {
	r.register(reflect.TypeOf((*T)(nil)).Elem(), service)
}

func GetService[T any](r *ServiceRegistry) T {
	serviceType := reflect.TypeOf((*T)(nil)).Elem()
	service := r.get(serviceType)
	return service.(T)
}
