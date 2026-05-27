package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// ParseGoSource extracts package, struct, interface, and function symbols from a Go file.
func ParseGoSource(absPath string, rootPath string) (*Graph, error) {
	fset := token.NewFileSet()
	fileAST, err := parser.ParseFile(fset, absPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file %s: %w", absPath, err)
	}

	relPath, err := filepath.Rel(rootPath, absPath)
	if err != nil {
		relPath = filepath.Base(absPath)
	}
	relPath = filepath.ToSlash(relPath)

	g := NewGraph()
	pkgName := fileAST.Name.Name

	// 1. Create File Node
	g.AddNode(Node{
		ID:   relPath,
		Type: "File",
		Properties: map[string]interface{}{
			"filepath": relPath,
			"package":  pkgName,
		},
	})

	// 2. Parse Imports and build IMPORTS edges
	for _, impSpec := range fileAST.Imports {
		importPath := strings.Trim(impSpec.Path.Value, `"`)
		
		// We only model import edges for external/internal library packages
		importNodeID := "package:" + importPath
		g.AddNode(Node{
			ID:   importNodeID,
			Type: "Module", // Represents an imported package/module dependency
			Properties: map[string]interface{}{
				"path": importPath,
			},
		})

		g.AddEdge(relPath, importNodeID, "IMPORTS")
	}

	// Helper to get receiver type name for methods
	getReceiverTypeName := func(recv *ast.FieldList) string {
		if recv == nil || len(recv.List) == 0 {
			return ""
		}
		typeExpr := recv.List[0].Type
		// Handle pointer receivers (*Scheduler)
		if star, ok := typeExpr.(*ast.StarExpr); ok {
			typeExpr = star.X
		}
		if ident, ok := typeExpr.(*ast.Ident); ok {
			return ident.Name
		}
		return ""
	}

	// 3. Scan declarations for Structs, Interfaces, and Functions/Methods
	for _, decl := range fileAST.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			// General declarations: import, const, type, var
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					symbolName := typeSpec.Name.Name
					symbolID := pkgName + "." + symbolName
					symbolSubtype := "Type"

					switch typeSpec.Type.(type) {
					case *ast.StructType:
						symbolSubtype = "Struct"
					case *ast.InterfaceType:
						symbolSubtype = "Interface"
					}

					g.AddNode(Node{
						ID:   symbolID,
						Type: "Symbol",
						Properties: map[string]interface{}{
							"name":     symbolName,
							"subtype":  symbolSubtype,
							"package":  pkgName,
							"filepath": relPath,
						},
					})

					// (Symbol) -[DEFINED_IN]-> (File)
					g.AddEdge(symbolID, relPath, "DEFINED_IN")
				}
			}

		case *ast.FuncDecl:
			// Function or Method declaration
			funcName := d.Name.Name
			receiverType := getReceiverTypeName(d.Recv)

			var symbolID, symbolSubtype string
			if receiverType != "" {
				symbolSubtype = "Method"
				symbolID = pkgName + "." + receiverType + "." + funcName
			} else {
				symbolSubtype = "Function"
				symbolID = pkgName + "." + funcName
			}

			properties := map[string]interface{}{
				"name":     funcName,
				"subtype":  symbolSubtype,
				"package":  pkgName,
				"filepath": relPath,
			}
			if receiverType != "" {
				properties["receiver"] = receiverType
			}

			g.AddNode(Node{
				ID:         symbolID,
				Type:       "Symbol",
				Properties: properties,
			})

			// (Symbol) -[DEFINED_IN]-> (File)
			g.AddEdge(symbolID, relPath, "DEFINED_IN")
		}
	}

	return g, nil
}
