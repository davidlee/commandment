package nodemanager

// DisplayNodeTreeCommandParams contains parameters for displaying node trees.
type DisplayNodeTreeCommandParams struct {
	RootReference string
	MaxDepth      int
}

// NodeTree represents a tree structure of nodes with statistics.
type NodeTree struct {
	Nodes []Node
	Stats TreeStats
}

// CreateListCommandParams contains parameters for creating lists.
type CreateListCommandParams struct {
	Title       string
	Description string
	ParentID    *int64
}

// NodeCommandResult represents the result of node operations.
type NodeCommandResult struct {
	Node   Node
	Errors []ValidationError
}

// ShowNodeQueryParams contains parameters for querying individual nodes.
type ShowNodeQueryParams struct {
	Ref int64
}

// Node represents a domain object for nodes in the system.
type Node struct {
	ID          int64
	Title       string
	Description string
}

// TreeStats contains statistics about a tree structure.
type TreeStats struct {
	TotalNodes int
	MaxDepth   int
}

// ValidationError represents validation errors.
type ValidationError struct {
	Field   string
	Message string
}
