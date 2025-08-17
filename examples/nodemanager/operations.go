package nodemanager

import (
	"context"

	"github.com/davidlee/commandment/pkg/commandment"
)

// ShowNodeQuery implements a read-only query for retrieving individual nodes.
type ShowNodeQuery struct {
	Params  ShowNodeQueryParams
	Service NodeService
	Meta    commandment.OperationMetadata
	Logger  commandment.Logger
}

func (q *ShowNodeQuery) Execute(ctx context.Context) (Node, error) {
	return commandment.ExecuteOperation(ctx, q, func(ctx context.Context) (Node, error) {
		return q.Service.ShowNode(ctx, q.Params)
	})
}

func (q *ShowNodeQuery) Metadata() commandment.OperationMetadata {
	return q.Meta
}

func (q *ShowNodeQuery) Descriptor() commandment.OperationDescriptor {
	return commandment.OperationDescriptor{
		Type:     "ShowNodeQuery",
		Params:   q.Params,
		Metadata: q.Meta,
	}
}

func (q *ShowNodeQuery) GetMetadata() *commandment.OperationMetadata { return &q.Meta }
func (q *ShowNodeQuery) GetLogger() commandment.Logger               { return q.Logger }

// DisplayNodeTreeCommand implements a command for displaying node trees (updates node refs).
type DisplayNodeTreeCommand struct {
	Params  DisplayNodeTreeCommandParams
	Service TreeService
	Meta    commandment.OperationMetadata
	Logger  commandment.Logger
}

func (c *DisplayNodeTreeCommand) Execute(ctx context.Context) (NodeTree, error) {
	return commandment.ExecuteOperation(ctx, c, func(ctx context.Context) (NodeTree, error) {
		return c.Service.DisplayTree(ctx, c.Params)
	})
}

func (c *DisplayNodeTreeCommand) Metadata() commandment.OperationMetadata {
	return c.Meta
}

func (c *DisplayNodeTreeCommand) Descriptor() commandment.OperationDescriptor {
	return commandment.OperationDescriptor{
		Type:     "DisplayNodeTreeCommand",
		Params:   c.Params,
		Metadata: c.Meta,
	}
}

func (c *DisplayNodeTreeCommand) GetMetadata() *commandment.OperationMetadata { return &c.Meta }
func (c *DisplayNodeTreeCommand) GetLogger() commandment.Logger               { return c.Logger }

// CreateListCommand implements a command for creating lists (mutates state).
type CreateListCommand struct {
	Params  CreateListCommandParams
	Service ListService
	Meta    commandment.OperationMetadata
	Logger  commandment.Logger
}

func (c *CreateListCommand) Execute(ctx context.Context) (NodeCommandResult, error) {
	return commandment.ExecuteOperation(ctx, c, func(ctx context.Context) (NodeCommandResult, error) {
		return c.Service.CreateList(ctx, c.Params)
	})
}

func (c *CreateListCommand) Metadata() commandment.OperationMetadata {
	return c.Meta
}

func (c *CreateListCommand) Descriptor() commandment.OperationDescriptor {
	return commandment.OperationDescriptor{
		Type:     "CreateListCommand",
		Params:   c.Params,
		Metadata: c.Meta,
	}
}

func (c *CreateListCommand) GetMetadata() *commandment.OperationMetadata { return &c.Meta }
func (c *CreateListCommand) GetLogger() commandment.Logger               { return c.Logger }
