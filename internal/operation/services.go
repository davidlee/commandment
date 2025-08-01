package operation

import "context"

// Services remain focused on business logic
type TreeService interface {
	DisplayTree(ctx context.Context, params DisplayNodeTreeCommandParams) (NodeTree, error)
}

type ListService interface {
	CreateList(ctx context.Context, params CreateListCommandParams) (NodeCommandResult, error)
}

type NodeService interface {
	ShowNode(ctx context.Context, params ShowNodeQueryParams) (Node, error)
}
