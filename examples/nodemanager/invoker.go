package nodemanager

import "com.github/davidlee/commandment/pkg/operation"

// QueryInvoker provides methods for creating read-only query operations.
type QueryInvoker interface {
	NewShowNodeQuery(params ShowNodeQueryParams) (*ShowNodeQuery, error)
}

// CommandInvoker provides methods for creating command operations that mutate state.
type CommandInvoker interface {
	NewDisplayNodeTreeCommand(params DisplayNodeTreeCommandParams) (*DisplayNodeTreeCommand, error)
	NewCreateListCommand(params CreateListCommandParams) (*CreateListCommand, error)
}

// OperationInvoker combines QueryInvoker and CommandInvoker for full operation creation capabilities.
type OperationInvoker interface {
	QueryInvoker
	CommandInvoker
}

// NodeManagerBus wraps the operation framework bus and provides domain-specific operation creation.
type NodeManagerBus struct {
	bus *operation.OperationBus
}

// NewNodeManagerBus creates a new NodeManagerBus wrapping the operation framework.
func NewNodeManagerBus(bus *operation.OperationBus) *NodeManagerBus {
	return &NodeManagerBus{bus: bus}
}

// NewShowNodeQuery creates a new ShowNodeQuery operation.
func (b *NodeManagerBus) NewShowNodeQuery(params ShowNodeQueryParams) (*ShowNodeQuery, error) {
	return operation.CreateOperation[*ShowNodeQuery](b.bus, params)
}

// NewDisplayNodeTreeCommand creates a new DisplayNodeTreeCommand operation.
func (b *NodeManagerBus) NewDisplayNodeTreeCommand(params DisplayNodeTreeCommandParams) (*DisplayNodeTreeCommand, error) {
	return operation.CreateOperation[*DisplayNodeTreeCommand](b.bus, params)
}

// NewCreateListCommand creates a new CreateListCommand operation.
func (b *NodeManagerBus) NewCreateListCommand(params CreateListCommandParams) (*CreateListCommand, error) {
	return operation.CreateOperation[*CreateListCommand](b.bus, params)
}