package account

import (
	"fmt"
	"sort"
	"sync"
)

// backendRegistry is the global registry of all registered backends.
// It is initialized at package load time and populated by init() functions.
var backendRegistry = &registry{
	backends: make(map[string]Backend),
}

// registry maintains the mapping of backend kinds to their implementations.
// It is thread-safe for concurrent reads and writes.
type registry struct {
	mu       sync.RWMutex
	backends map[string]Backend
}

// RegisterBackend registers a backend implementation.
//
// This is typically called from init() functions in backend implementation files:
//
//	func init() {
//	    account.RegisterBackend(&myBackend{})
//	}
//
// Panics if a backend with the same kind is already registered.
// This ensures duplicate registrations are caught early during initialization.
func RegisterBackend(b Backend) {
	backendRegistry.mu.Lock()
	defer backendRegistry.mu.Unlock()

	kind := b.Kind()
	if kind == "" {
		panic("backend kind cannot be empty")
	}

	if _, exists := backendRegistry.backends[kind]; exists {
		panic(fmt.Sprintf("backend %q already registered", kind))
	}

	backendRegistry.backends[kind] = b
}

// GetBackend retrieves a registered backend by its kind.
// Returns the backend and true if found, or nil and false if not found.
func GetBackend(kind string) (Backend, bool) {
	backendRegistry.mu.RLock()
	defer backendRegistry.mu.RUnlock()

	b, ok := backendRegistry.backends[kind]
	return b, ok
}

// ListBackends returns all registered backend kinds in sorted order.
// This is useful for discovery and documentation.
func ListBackends() []string {
	backendRegistry.mu.RLock()
	defer backendRegistry.mu.RUnlock()

	kinds := make([]string, 0, len(backendRegistry.backends))
	for kind := range backendRegistry.backends {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return kinds
}

// IsRegisteredBackend checks if a backend kind is registered.
func IsRegisteredBackend(kind string) bool {
	_, ok := GetBackend(kind)
	return ok
}

// GetBackendMetadata retrieves metadata for a registered backend.
// Returns the metadata and true if found, or empty metadata and false if not found.
func GetBackendMetadata(kind string) (BackendMetadata, bool) {
	b, ok := GetBackend(kind)
	if !ok {
		return BackendMetadata{}, false
	}
	return b.Metadata(), true
}

// ListBackendsWithMetadata returns all registered backends with their metadata.
// This is useful for building dynamic UIs or documentation.
func ListBackendsWithMetadata() map[string]BackendMetadata {
	backendRegistry.mu.RLock()
	defer backendRegistry.mu.RUnlock()

	result := make(map[string]BackendMetadata, len(backendRegistry.backends))
	for kind, b := range backendRegistry.backends {
		result[kind] = b.Metadata()
	}
	return result
}

// unregisterBackend removes a backend from the registry.
// This is primarily intended for testing and should not be used in production code.
func unregisterBackend(kind string) {
	backendRegistry.mu.Lock()
	defer backendRegistry.mu.Unlock()
	delete(backendRegistry.backends, kind)
}

// clearRegistry removes all backends from the registry.
// This is intended only for testing purposes.
func clearRegistry() {
	backendRegistry.mu.Lock()
	defer backendRegistry.mu.Unlock()
	backendRegistry.backends = make(map[string]Backend)
}
