package operation

// DisplayNodeTreeCommand parameters and results
type DisplayNodeTreeCommandParams struct {
	RootReference string
	MaxDepth      int
}

type NodeTree struct {
	Nodes []Node
	Stats TreeStats
}

// CreateListCommand parameters and results
type CreateListCommandParams struct {
	Title       string
	Description string
	ParentID    *int64
}

type NodeCommandResult struct {
	Node   Node
	Errors []ValidationError
}

// ShowNodeQuery parameters and results
type ShowNodeQueryParams struct {
	Ref int64
}

// Domain objects
type Node struct {
	ID          int64
	Title       string
	Description string
	// Add other node fields as needed
}

type TreeStats struct {
	TotalNodes int
	MaxDepth   int
}

type ValidationError struct {
	Field   string
	Message string
}
