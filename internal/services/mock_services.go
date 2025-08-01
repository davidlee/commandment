package services

import (
	"context"
	"fmt"

	"com.github/davidlee/commandment/internal/operation"
)

// Mock TreeService implementation
type MockTreeService struct{}

func NewMockTreeService() *MockTreeService {
	return &MockTreeService{}
}

func (s *MockTreeService) DisplayTree(ctx context.Context, params operation.DisplayNodeTreeCommandParams) (operation.NodeTree, error) {
	// Mock implementation - would normally query database and update refs
	nodes := []operation.Node{
		{ID: 1, Title: "Root Node", Description: "The root of the tree"},
		{ID: 2, Title: "Child Node 1", Description: "First child"},
		{ID: 3, Title: "Child Node 2", Description: "Second child"},
	}

	if params.MaxDepth <= 0 {
		return operation.NodeTree{}, fmt.Errorf("MaxDepth must be positive, got %d", params.MaxDepth)
	}

	return operation.NodeTree{
		Nodes: nodes[:min(len(nodes), params.MaxDepth+1)],
		Stats: operation.TreeStats{
			TotalNodes: len(nodes),
			MaxDepth:   2,
		},
	}, nil
}

// Mock ListService implementation
type MockListService struct{}

func NewMockListService() *MockListService {
	return &MockListService{}
}

func (s *MockListService) CreateList(ctx context.Context, params operation.CreateListCommandParams) (operation.NodeCommandResult, error) {
	// Mock implementation - would normally create list in database
	if params.Title == "" {
		return operation.NodeCommandResult{
			Errors: []operation.ValidationError{
				{Field: "Title", Message: "Title is required"},
			},
		}, nil
	}

	node := operation.Node{
		ID:          42, // Mock generated ID
		Title:       params.Title,
		Description: params.Description,
	}

	return operation.NodeCommandResult{
		Node:   node,
		Errors: nil,
	}, nil
}

// Mock NodeService implementation
type MockNodeService struct{}

func NewMockNodeService() *MockNodeService {
	return &MockNodeService{}
}

func (s *MockNodeService) ShowNode(ctx context.Context, params operation.ShowNodeQueryParams) (operation.Node, error) {
	// Mock implementation - would normally query database
	if params.Ref <= 0 {
		return operation.Node{}, fmt.Errorf("invalid node reference: %d", params.Ref)
	}

	return operation.Node{
		ID:          params.Ref,
		Title:       fmt.Sprintf("Node %d", params.Ref),
		Description: fmt.Sprintf("This is node with ID %d", params.Ref),
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
