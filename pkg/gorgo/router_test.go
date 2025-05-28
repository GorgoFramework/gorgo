package gorgo

import (
	"testing"
)

func TestRouterParameterExtraction(t *testing.T) {
	router := NewRouter()

	// Add a route with parameters
	handler := func(ctx *Context) error {
		return nil
	}
	router.AddRoute("GET", "/users/:id/posts/:postId", handler)

	// Test parameter extraction
	foundHandler, params := router.FindHandler("GET", "/users/123/posts/456")

	if foundHandler == nil {
		t.Error("Expected to find handler, got nil")
	}

	if params == nil {
		t.Error("Expected to get parameters, got nil")
	}

	if params["id"] != "123" {
		t.Errorf("Expected id parameter to be '123', got '%s'", params["id"])
	}

	if params["postId"] != "456" {
		t.Errorf("Expected postId parameter to be '456', got '%s'", params["postId"])
	}
}

func TestRouterExactMatch(t *testing.T) {
	router := NewRouter()

	handler := func(ctx *Context) error {
		return nil
	}
	router.AddRoute("GET", "/users/profile", handler)

	// Test exact match (should return nil params)
	foundHandler, params := router.FindHandler("GET", "/users/profile")

	if foundHandler == nil {
		t.Error("Expected to find handler, got nil")
	}

	if params != nil && len(params) > 0 {
		t.Error("Expected no parameters for exact match, got some")
	}
}

func TestRouterNoMatch(t *testing.T) {
	router := NewRouter()

	handler := func(ctx *Context) error {
		return nil
	}
	router.AddRoute("GET", "/users/:id", handler)

	// Test no match
	foundHandler, params := router.FindHandler("GET", "/posts/123")

	if foundHandler != nil {
		t.Error("Expected no handler, got one")
	}

	if params != nil {
		t.Error("Expected no parameters, got some")
	}
}
