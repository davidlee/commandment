package nodemanager

import (
	"context"

	"com.github/davidlee/commandment/pkg/operation"
)

// ShowNodeQuery implements a read-only query for retrieving individual nodes.
type ShowNodeQuery struct {
	Params  ShowNodeQueryParams
	Service NodeService
	Meta    operation.OperationMetadata
	Logger  operation.Logger
}

func (q *ShowNodeQuery) Execute(ctx context.Context) (Node, error) {
	return operation.ExecuteOperation(q, func() (Node, error) {
		return q.Service.ShowNode(ctx, q.Params)
	})
}

func (q *ShowNodeQuery) Metadata() operation.OperationMetadata {
	return q.Meta
}

func (q *ShowNodeQuery) Descriptor() operation.OperationDescriptor {
	return operation.OperationDescriptor{
		Type:     "ShowNodeQuery",
		Params:   q.Params,
		Metadata: q.Meta,
	}
}

func (q *ShowNodeQuery) GetMetadata() *operation.OperationMetadata { return &q.Meta }
func (q *ShowNodeQuery) GetLogger() operation.Logger               { return q.Logger }

// DisplayNodeTreeCommand implements a command for displaying node trees (updates node refs).
type DisplayNodeTreeCommand struct {
	Params  DisplayNodeTreeCommandParams
	Service TreeService
	Meta    operation.OperationMetadata
	Logger  operation.Logger
}

func (c *DisplayNodeTreeCommand) Execute(ctx context.Context) (NodeTree, error) {
	return operation.ExecuteOperation(c, func() (NodeTree, error) {
		return c.Service.DisplayTree(ctx, c.Params)
	})
}

func (c *DisplayNodeTreeCommand) Metadata() operation.OperationMetadata {
	return c.Meta
}

func (c *DisplayNodeTreeCommand) Descriptor() operation.OperationDescriptor {
	return operation.OperationDescriptor{
		Type:     "DisplayNodeTreeCommand",
		Params:   c.Params,
		Metadata: c.Meta,
	}
}

func (c *DisplayNodeTreeCommand) GetMetadata() *operation.OperationMetadata { return &c.Meta }
func (c *DisplayNodeTreeCommand) GetLogger() operation.Logger               { return c.Logger }

// CreateListCommand implements a command for creating lists (mutates state).
type CreateListCommand struct {
	Params  CreateListCommandParams
	Service ListService
	Meta    operation.OperationMetadata
	Logger  operation.Logger
}

func (c *CreateListCommand) Execute(ctx context.Context) (NodeCommandResult, error) {
	return operation.ExecuteOperation(c, func() (NodeCommandResult, error) {
		return c.Service.CreateList(ctx, c.Params)
	})
}

func (c *CreateListCommand) Metadata() operation.OperationMetadata {
	return c.Meta
}

func (c *CreateListCommand) Descriptor() operation.OperationDescriptor {
	return operation.OperationDescriptor{
		Type:     "CreateListCommand",
		Params:   c.Params,
		Metadata: c.Meta,
	}
}

func (c *CreateListCommand) GetMetadata() *operation.OperationMetadata { return &c.Meta }
func (c *CreateListCommand) GetLogger() operation.Logger               { return c.Logger }