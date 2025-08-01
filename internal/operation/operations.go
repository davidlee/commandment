package operation

import "context"

// Concrete query implementation (read-only)
type ShowNodeQuery struct {
	Params        ShowNodeQueryParams
	Service       NodeService
	OperationMeta OperationMetadata
	Logger        Logger
}

func (q *ShowNodeQuery) Execute(ctx context.Context) (Node, error) {
	return executeOperation(q, func() (Node, error) {
		return q.Service.ShowNode(ctx, q.Params)
	})
}

func (q *ShowNodeQuery) Metadata() OperationMetadata {
	return q.OperationMeta
}

func (q *ShowNodeQuery) Descriptor() OperationDescriptor {
	return OperationDescriptor{
		Type:     "ShowNodeQuery",
		Params:   q.Params,
		Metadata: q.OperationMeta,
	}
}

func (q *ShowNodeQuery) getMetadata() *OperationMetadata { return &q.OperationMeta }
func (q *ShowNodeQuery) getLogger() Logger               { return q.Logger }

// Concrete command implementation (updates node refs)
type DisplayNodeTreeCommand struct {
	Params        DisplayNodeTreeCommandParams
	Service       TreeService
	OperationMeta OperationMetadata
	Logger        Logger
}

func (c *DisplayNodeTreeCommand) Execute(ctx context.Context) (NodeTree, error) {
	return executeOperation(c, func() (NodeTree, error) {
		return c.Service.DisplayTree(ctx, c.Params)
	})
}

func (c *DisplayNodeTreeCommand) Metadata() OperationMetadata {
	return c.OperationMeta
}

func (c *DisplayNodeTreeCommand) Descriptor() OperationDescriptor {
	return OperationDescriptor{
		Type:     "DisplayNodeTreeCommand",
		Params:   c.Params,
		Metadata: c.OperationMeta,
	}
}

func (c *DisplayNodeTreeCommand) getMetadata() *OperationMetadata { return &c.OperationMeta }
func (c *DisplayNodeTreeCommand) getLogger() Logger               { return c.Logger }

// Concrete command implementation (mutates state)
type CreateListCommand struct {
	Params        CreateListCommandParams
	Service       ListService
	OperationMeta OperationMetadata
	Logger        Logger
}

func (c *CreateListCommand) Execute(ctx context.Context) (NodeCommandResult, error) {
	return executeOperation(c, func() (NodeCommandResult, error) {
		return c.Service.CreateList(ctx, c.Params)
	})
}

func (c *CreateListCommand) Metadata() OperationMetadata {
	return c.OperationMeta
}

func (c *CreateListCommand) Descriptor() OperationDescriptor {
	return OperationDescriptor{
		Type:     "CreateListCommand",
		Params:   c.Params,
		Metadata: c.OperationMeta,
	}
}

func (c *CreateListCommand) getMetadata() *OperationMetadata { return &c.OperationMeta }
func (c *CreateListCommand) getLogger() Logger               { return c.Logger }
