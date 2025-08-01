package operation

import "context"

// TreeService provides operations for managing and displaying node trees.
type TreeService interface {
	DisplayTree(ctx context.Context, params DisplayNodeTreeCommandParams) (NodeTree, error)
}

// ListService provides operations for creating and managing lists.
type ListService interface {
	CreateList(ctx context.Context, params CreateListCommandParams) (NodeCommandResult, error)
}

// NodeService provides operations for retrieving individual nodes.
type NodeService interface {
	ShowNode(ctx context.Context, params ShowNodeQueryParams) (Node, error)
}
