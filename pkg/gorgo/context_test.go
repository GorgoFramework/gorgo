package gorgo

import (
	"testing"

	"github.com/GorgoFramework/gorgo/internal/container"
	"github.com/valyala/fasthttp"
)

func TestContextParamMethods(t *testing.T) {
	// Create a mock fasthttp context
	ctx := &fasthttp.RequestCtx{}
	container := container.NewContainer()
	plugins := make(map[string]Plugin)

	// Create Gorgo context
	gorgoCtx := NewContext(ctx, container, plugins)

	// Test SetParam and Param
	gorgoCtx.SetParam("name", "John")
	gorgoCtx.SetParam("age", "25")

	if gorgoCtx.Param("name") != "John" {
		t.Errorf("Expected name to be 'John', got '%s'", gorgoCtx.Param("name"))
	}

	if gorgoCtx.Param("age") != "25" {
		t.Errorf("Expected age to be '25', got '%s'", gorgoCtx.Param("age"))
	}

	// Test ParamDefault
	role := gorgoCtx.ParamDefault("role", "user")
	if role != "user" {
		t.Errorf("Expected default role to be 'user', got '%s'", role)
	}

	name := gorgoCtx.ParamDefault("name", "Anonymous")
	if name != "John" {
		t.Errorf("Expected existing name to be 'John', got '%s'", name)
	}

	// Test HasParam
	if !gorgoCtx.HasParam("name") {
		t.Error("Expected HasParam('name') to be true")
	}

	if gorgoCtx.HasParam("nonexistent") {
		t.Error("Expected HasParam('nonexistent') to be false")
	}

	// Test Params
	allParams := gorgoCtx.Params()
	if len(allParams) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(allParams))
	}

	if allParams["name"] != "John" {
		t.Errorf("Expected name in params to be 'John', got '%s'", allParams["name"])
	}

	if allParams["age"] != "25" {
		t.Errorf("Expected age in params to be '25', got '%s'", allParams["age"])
	}

	// Test that Params returns a copy (modification shouldn't affect original)
	allParams["name"] = "Modified"
	if gorgoCtx.Param("name") != "John" {
		t.Error("Modifying returned params map should not affect original parameters")
	}
}
