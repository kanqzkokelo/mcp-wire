package protocol

import (
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p/core/network"
)

type ServiceHandler func(stream network.Stream)

type ServiceRegistry struct {
	mu       sync.RWMutex
	services map[string]ServiceHandler
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]ServiceHandler),
	}
}

func (r *ServiceRegistry) Register(serviceName string, handler ServiceHandler) error {
	if err := ValidateServiceName(serviceName); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.services[serviceName]; exists {
		return fmt.Errorf("service '%s' is already registered", serviceName)
	}
	r.services[serviceName] = handler
	return nil
}

func (r *ServiceRegistry) Get(serviceName string) (ServiceHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.services[serviceName]
	return handler, ok
}

func (r *ServiceRegistry) Unregister(serviceName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.services, serviceName)
}

func (r *ServiceRegistry) ListServices() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	return names
}
