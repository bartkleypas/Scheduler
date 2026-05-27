package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGoSource(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "code_parser_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockSourceContent := `package demo

import (
	"context"
	"fmt"
)

type Config struct {
	Port int
}

type Runner interface {
	Run(ctx context.Context) error
}

func NewConfig() *Config {
	return &Config{Port: 8080}
}

func (c *Config) Start() error {
	fmt.Println("Starting...")
	return nil
}
`

	mockFilePath := filepath.Join(tempDir, "scheduler.go")
	err = os.WriteFile(mockFilePath, []byte(mockSourceContent), 0644)
	if err != nil {
		t.Fatalf("failed to write mock source file: %v", err)
	}

	graph, err := ParseGoSource(mockFilePath, tempDir)
	if err != nil {
		t.Fatalf("ParseGoSource failed: %v", err)
	}

	// 1. Verify File Node
	fileNode, ok := graph.Nodes["scheduler.go"]
	if !ok {
		t.Fatal("expected File node 'scheduler.go' not found")
	}
	if fileNode.Type != "File" {
		t.Errorf("expected File node type 'File'; got %q", fileNode.Type)
	}
	if fileNode.Properties["package"] != "demo" {
		t.Errorf("expected File package 'demo'; got %q", fileNode.Properties["package"])
	}

	// 2. Verify Imports are mapped
	importNode, ok := graph.Nodes["package:context"]
	if !ok {
		t.Fatal("expected imported package node 'package:context' not found")
	}
	if importNode.Type != "Module" {
		t.Errorf("expected import node type 'Module'; got %q", importNode.Type)
	}

	// Verify Import Edge
	hasImportEdge := false
	for _, edge := range graph.Edges {
		if edge.From == "scheduler.go" && edge.To == "package:context" && edge.Type == "IMPORTS" {
			hasImportEdge = true
		}
	}
	if !hasImportEdge {
		t.Error("expected IMPORTS edge was not found")
	}

	// 3. Verify Struct and Interface are mapped
	structNode, ok := graph.Nodes["demo.Config"]
	if !ok {
		t.Fatal("expected struct node 'demo.Config' not found")
	}
	if structNode.Properties["subtype"] != "Struct" {
		t.Errorf("expected Config subtype 'Struct'; got %q", structNode.Properties["subtype"])
	}

	interfaceNode, ok := graph.Nodes["demo.Runner"]
	if !ok {
		t.Fatal("expected interface node 'demo.Runner' not found")
	}
	if interfaceNode.Properties["subtype"] != "Interface" {
		t.Errorf("expected Runner subtype 'Interface'; got %q", interfaceNode.Properties["subtype"])
	}

	// 4. Verify Function and Method are mapped
	funcNode, ok := graph.Nodes["demo.NewConfig"]
	if !ok {
		t.Fatal("expected function node 'demo.NewConfig' not found")
	}
	if funcNode.Properties["subtype"] != "Function" {
		t.Errorf("expected NewConfig subtype 'Function'; got %q", funcNode.Properties["subtype"])
	}

	methodNode, ok := graph.Nodes["demo.Config.Start"]
	if !ok {
		t.Fatal("expected method node 'demo.Config.Start' not found")
	}
	if methodNode.Properties["subtype"] != "Method" {
		t.Errorf("expected Start subtype 'Method'; got %q", methodNode.Properties["subtype"])
	}
	if methodNode.Properties["receiver"] != "Config" {
		t.Errorf("expected method receiver 'Config'; got %q", methodNode.Properties["receiver"])
	}

	// Verify DEFINED_IN Edges exist for symbols
	definedInEdgesCount := 0
	for _, edge := range graph.Edges {
		if edge.To == "scheduler.go" && edge.Type == "DEFINED_IN" {
			definedInEdgesCount++
		}
	}
	if definedInEdgesCount != 4 {
		t.Errorf("expected 4 DEFINED_IN edges linking symbols to scheduler.go; got %d", definedInEdgesCount)
	}
}
