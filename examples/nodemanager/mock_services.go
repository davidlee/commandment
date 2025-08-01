// Package nodemanager provides an example implementation of the operation pattern
// for node and tree management operations.
package nodemanager

import (
	"context"
	"fmt"
)

// MockTreeService provides a mock implementation of TreeService.
type MockTreeService struct{}

// NewMockTreeService creates a new MockTreeService.
func NewMockTreeService() *MockTreeService {
	return &MockTreeService{}
}

// DisplayTree implements TreeService.DisplayTree with mock behavior.
func (s *MockTreeService) DisplayTree(ctx context.Context, params DisplayNodeTreeCommandParams) (NodeTree, error) {
	nodes := []Node{
		{ID: 1, Title: "Root Node", Description: "The root of the tree"},
		{ID: 2, Title: "Child Node 1", Description: "First child"},
		{ID: 3, Title: "Child Node 2", Description: "Second child"},
	}

	if params.MaxDepth <= 0 {
		return NodeTree{}, fmt.Errorf("MaxDepth must be positive, got %d", params.MaxDepth)
	}

	return NodeTree{
		Nodes: nodes[:minInt(len(nodes), params.MaxDepth+1)],
		Stats: TreeStats{
			TotalNodes: len(nodes),
			MaxDepth:   2,
		},
	}, nil
}

// MockListService provides a mock implementation of ListService.
type MockListService struct{}

// NewMockListService creates a new MockListService.
func NewMockListService() *MockListService {
	return &MockListService{}
}

// CreateList implements ListService.CreateList with mock behavior.
func (s *MockListService) CreateList(ctx context.Context, params CreateListCommandParams) (NodeCommandResult, error) {
	if params.Title == "" {
		return NodeCommandResult{
			Errors: []ValidationError{
				{Field: "Title", Message: "Title is required"},
			},
		}, nil
	}

	node := Node{
		ID:          42, // Mock generated ID
		Title:       params.Title,
		Description: params.Description,
	}

	return NodeCommandResult{
		Node:   node,
		Errors: nil,
	}, nil
}

// MockNodeService provides a mock implementation of NodeService.
type MockNodeService struct{}

// NewMockNodeService creates a new MockNodeService.
func NewMockNodeService() *MockNodeService {
	return &MockNodeService{}
}

// ShowNode implements NodeService.ShowNode with mock behavior.
func (s *MockNodeService) ShowNode(ctx context.Context, params ShowNodeQueryParams) (Node, error) {
	if params.Ref <= 0 {
		return Node{}, fmt.Errorf("invalid node reference: %d", params.Ref)
	}

	return Node{
		ID:          params.Ref,
		Title:       fmt.Sprintf("Node %d", params.Ref),
		Description: fmt.Sprintf("This is node with ID %d", params.Ref),
	}, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}