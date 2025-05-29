package container

import (
	"fmt"
	"testing"
)

// TestService - simple test service
type TestService struct {
	Name string
	ID   int
}

// TestInterface - interface for testing
type TestInterface interface {
	GetValue() string
}

// TestImplementation - implementation of the interface
type TestImplementation struct {
	value string
}

func (t *TestImplementation) GetValue() string {
	return t.value
}

func TestNewContainer(t *testing.T) {
	container := NewContainer()

	if container == nil {
		t.Fatal("NewContainer returned nil")
	}

	if container.services == nil {
		t.Fatal("services map is nil")
	}

	if len(container.services) != 0 {
		t.Fatal("services map should be empty initially")
	}
}

func TestContainer_Register(t *testing.T) {
	container := NewContainer()
	service := &TestService{Name: "test", ID: 42}

	container.Register("test-service", service)

	// Check that the service was registered
	retrievedService, exists := container.Get("test-service")
	if !exists {
		t.Fatal("service was not registered")
	}

	if retrievedService != service {
		t.Fatal("retrieved service is not the same as registered")
	}
}

func TestContainer_RegisterMultipleServices(t *testing.T) {
	container := NewContainer()

	service1 := &TestService{Name: "service1", ID: 1}
	service2 := &TestService{Name: "service2", ID: 2}

	container.Register("service1", service1)
	container.Register("service2", service2)

	// Check the first service
	retrieved1, exists1 := container.Get("service1")
	if !exists1 {
		t.Fatal("service1 was not found")
	}
	if retrieved1.(*TestService).ID != 1 {
		t.Fatal("service1 has wrong ID")
	}

	// Check the second service
	retrieved2, exists2 := container.Get("service2")
	if !exists2 {
		t.Fatal("service2 was not found")
	}
	if retrieved2.(*TestService).ID != 2 {
		t.Fatal("service2 has wrong ID")
	}
}

func TestContainer_RegisterOverwrite(t *testing.T) {
	container := NewContainer()

	service1 := &TestService{Name: "original", ID: 1}
	service2 := &TestService{Name: "updated", ID: 2}

	container.Register("service", service1)
	container.Register("service", service2) // Overwrite

	retrieved, exists := container.Get("service")
	if !exists {
		t.Fatal("service was not found")
	}

	retrievedService := retrieved.(*TestService)
	if retrievedService.Name != "updated" || retrievedService.ID != 2 {
		t.Fatal("service was not overwritten correctly")
	}
}

func TestContainer_Get(t *testing.T) {
	container := NewContainer()
	service := &TestService{Name: "test", ID: 42}

	container.Register("test-service", service)

	// Test getting existing service
	retrieved, exists := container.Get("test-service")
	if !exists {
		t.Fatal("existing service was not found")
	}
	if retrieved != service {
		t.Fatal("retrieved service is different from registered")
	}

	// Test getting non-existent service
	_, exists = container.Get("non-existent")
	if exists {
		t.Fatal("non-existent service was found")
	}
}

func TestContainer_GetTyped_Success(t *testing.T) {
	container := NewContainer()
	service := &TestService{Name: "test", ID: 42}

	container.Register("test-service", service)

	var retrieved *TestService
	err := container.GetTyped("test-service", &retrieved)

	if err != nil {
		t.Fatalf("GetTyped failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("retrieved service is nil")
	}

	if retrieved.Name != "test" || retrieved.ID != 42 {
		t.Fatal("retrieved service has wrong values")
	}
}

func TestContainer_GetTyped_ServiceNotFound(t *testing.T) {
	container := NewContainer()

	var retrieved *TestService
	err := container.GetTyped("non-existent", &retrieved)

	if err == nil {
		t.Fatal("GetTyped should have failed for non-existent service")
	}

	expectedError := "service non-existent not found"
	if err.Error() != expectedError {
		t.Fatalf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestContainer_GetTyped_NotAPointer(t *testing.T) {
	container := NewContainer()
	service := &TestService{Name: "test", ID: 42}

	container.Register("test-service", service)

	// Pass a non-pointer
	var retrieved TestService
	err := container.GetTyped("test-service", retrieved)

	if err == nil {
		t.Fatal("GetTyped should have failed for non-pointer target")
	}

	expectedError := "target is not a pointer"
	if err.Error() != expectedError {
		t.Fatalf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestContainer_GetTyped_WithInterface(t *testing.T) {
	container := NewContainer()
	implementation := &TestImplementation{value: "test-value"}

	container.Register("test-interface", implementation)

	var retrieved TestInterface
	err := container.GetTyped("test-interface", &retrieved)

	if err != nil {
		t.Fatalf("GetTyped failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("retrieved interface is nil")
	}

	if retrieved.GetValue() != "test-value" {
		t.Fatal("retrieved interface has wrong value")
	}
}

func TestContainer_GetTyped_WithPrimitiveTypes(t *testing.T) {
	container := NewContainer()

	// Test with int
	container.Register("int-value", 42)
	var intValue int
	err := container.GetTyped("int-value", &intValue)
	if err != nil {
		t.Fatalf("GetTyped failed for int: %v", err)
	}
	if intValue != 42 {
		t.Fatalf("expected 42, got %d", intValue)
	}

	// Test with string
	container.Register("string-value", "hello")
	var stringValue string
	err = container.GetTyped("string-value", &stringValue)
	if err != nil {
		t.Fatalf("GetTyped failed for string: %v", err)
	}
	if stringValue != "hello" {
		t.Fatalf("expected 'hello', got '%s'", stringValue)
	}
}

func TestContainer_ConcurrentAccess(t *testing.T) {
	container := NewContainer()

	// Test concurrent access
	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	// Concurrent registration
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			service := &TestService{Name: "concurrent", ID: id}
			container.Register(fmt.Sprintf("service-%d", id), service)
			done <- true
		}(i)
	}

	// Wait for registration completion
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Concurrent reading
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			service, exists := container.Get(fmt.Sprintf("service-%d", id))
			if !exists {
				t.Errorf("service-%d not found", id)
			}
			if service.(*TestService).ID != id {
				t.Errorf("service-%d has wrong ID", id)
			}
			done <- true
		}(i)
	}

	// Wait for reading completion
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// Benchmark for service registration
func BenchmarkContainer_Register(b *testing.B) {
	container := NewContainer()
	service := &TestService{Name: "benchmark", ID: 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		container.Register(fmt.Sprintf("service-%d", i), service)
	}
}

// Benchmark for getting services
func BenchmarkContainer_Get(b *testing.B) {
	container := NewContainer()
	service := &TestService{Name: "benchmark", ID: 1}
	container.Register("benchmark-service", service)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		container.Get("benchmark-service")
	}
}

// Benchmark for GetTyped
func BenchmarkContainer_GetTyped(b *testing.B) {
	container := NewContainer()
	service := &TestService{Name: "benchmark", ID: 1}
	container.Register("benchmark-service", service)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var retrieved *TestService
		container.GetTyped("benchmark-service", &retrieved)
	}
}
