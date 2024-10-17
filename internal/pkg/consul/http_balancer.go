package consul

import (
	"log"
	"sync"

	"github.com/hashicorp/consul/api"
)

type LeastConnectionsBalancer struct {
	mu          sync.RWMutex
	instances   []*api.ServiceEntry
	connections map[string]int
	errorNodes  map[string]struct{}
}

func NewLeastConnectionsBalancer(services []*api.ServiceEntry) *LeastConnectionsBalancer {
	balancer := &LeastConnectionsBalancer{
		instances:   services,
		connections: make(map[string]int),
		errorNodes:  make(map[string]struct{}),
	}

	return balancer
}

func (b *LeastConnectionsBalancer) RegisterBalancer(services []*api.ServiceEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.instances = services

	for _, instance := range services {
		log.Println(instance.Service.ID)
		b.connections[instance.Service.ID] = 0
	}
}

func (b *LeastConnectionsBalancer) MarkNodeAsError(instanceID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.errorNodes[instanceID] = struct{}{}
}

// ClearErrorNode clears the error status for the specified node.
func (b *LeastConnectionsBalancer) ClearErrorNode(instanceID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.errorNodes, instanceID)
}

// GetNextInstance selects the instance with the least connections and retries if necessary.
func (b *LeastConnectionsBalancer) GetNextInstance() *api.ServiceEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.instances) == 0 {
		return nil
	}

	// Find the instance with the least connections excluding error-marked nodes
	var minInstance *api.ServiceEntry
	minConnections := -1

	for _, instance := range b.instances {
		// Skip error-marked nodes
		if _, isErrorNode := b.errorNodes[instance.Service.ID]; isErrorNode {
			continue
		}

		connections, exists := b.connections[instance.Service.ID]
		if !exists || minConnections == -1 || connections < minConnections {
			minInstance = instance
			minConnections = connections
		}
	}

	// If no available instance found, return nil
	if minInstance == nil {
		return nil
	}

	// Increment the connection count for the selected instance
	b.connections[minInstance.Service.ID]++

	return minInstance
}

// ReleaseInstance decrements the connection count for the released instance.
func (b *LeastConnectionsBalancer) ReleaseInstance(instance *api.ServiceEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Decrement the connection count for the released instance
	if instance != nil {
		b.connections[instance.Service.ID]--
	}
}
