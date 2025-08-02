package nodemanager

import "github.com/davidlee/commandment/pkg/commandment"

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
	bus *commandment.OperationBus
}

// NewNodeManagerBus creates a new NodeManagerBus wrapping the operation framework.
func NewNodeManagerBus(bus *commandment.OperationBus) *NodeManagerBus {
	return &NodeManagerBus{bus: bus}
}

// NewShowNodeQuery creates a new ShowNodeQuery commandment.
func (b *NodeManagerBus) NewShowNodeQuery(params ShowNodeQueryParams) (*ShowNodeQuery, error) {
	return commandment.CreateOperation[*ShowNodeQuery](b.bus, params)
}

// NewDisplayNodeTreeCommand creates a new DisplayNodeTreeCommand commandment.
func (b *NodeManagerBus) NewDisplayNodeTreeCommand(params DisplayNodeTreeCommandParams) (*DisplayNodeTreeCommand, error) {
	return commandment.CreateOperation[*DisplayNodeTreeCommand](b.bus, params)
}

// NewCreateListCommand creates a new CreateListCommand commandment.
func (b *NodeManagerBus) NewCreateListCommand(params CreateListCommandParams) (*CreateListCommand, error) {
	return commandment.CreateOperation[*CreateListCommand](b.bus, params)
}
