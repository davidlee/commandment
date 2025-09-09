package commandment_test

import (
  "reflect"
  "testing"

  "github.com/davidlee/commandment/pkg/commandment"
)

type RegistryTestService struct {
  Name string
}

type DatabaseService struct {
  ConnectionString string
}

func TestGetServiceByType(t *testing.T) {
  registry := commandment.NewServiceRegistry()

  // Register test services
  testSvc := RegistryTestService{Name: "test"}
  dbSvc := DatabaseService{ConnectionString: "localhost:5432"}

  commandment.RegisterService(registry, testSvc)
  commandment.RegisterService(registry, dbSvc)

  // Test retrieving by type
  testType := reflect.TypeOf((*RegistryTestService)(nil)).Elem()
  retrieved := registry.GetServiceByType(testType)

  if retrieved == nil {
    t.Fatal("Expected service, got nil")
  }

  svc, ok := retrieved.(RegistryTestService)
  if !ok {
    t.Fatalf("Expected RegistryTestService, got %T", retrieved)
  }

  if svc.Name != "test" {
    t.Errorf("Expected Name 'test', got %q", svc.Name)
  }
}

func TestGetServiceByType_NotRegistered(t *testing.T) {
  registry := commandment.NewServiceRegistry()

  // Test retrieving unregistered service type
  testType := reflect.TypeOf((*RegistryTestService)(nil)).Elem()

  defer func() {
    if r := recover(); r == nil {
      t.Error("Expected panic for unregistered service type")
    }
  }()

  registry.GetServiceByType(testType)
}

func TestGetServiceByType_MultipleServices(t *testing.T) {
  registry := commandment.NewServiceRegistry()

  // Register multiple services
  testSvc := RegistryTestService{Name: "primary"}
  dbSvc := DatabaseService{ConnectionString: "prod:5432"}

  commandment.RegisterService(registry, testSvc)
  commandment.RegisterService(registry, dbSvc)

  // Test retrieving each service type
  testType := reflect.TypeOf((*RegistryTestService)(nil)).Elem()
  dbType := reflect.TypeOf((*DatabaseService)(nil)).Elem()

  retrievedTest := registry.GetServiceByType(testType)
  retrievedDb := registry.GetServiceByType(dbType)

  // Verify RegistryTestService
  testResult, ok := retrievedTest.(RegistryTestService)
  if !ok {
    t.Fatalf("Expected RegistryTestService, got %T", retrievedTest)
  }
  if testResult.Name != "primary" {
    t.Errorf("Expected Name 'primary', got %q", testResult.Name)
  }

  // Verify DatabaseService
  dbResult, ok := retrievedDb.(DatabaseService)
  if !ok {
    t.Fatalf("Expected DatabaseService, got %T", retrievedDb)
  }
  if dbResult.ConnectionString != "prod:5432" {
    t.Errorf("Expected ConnectionString 'prod:5432', got %q", dbResult.ConnectionString)
  }
}