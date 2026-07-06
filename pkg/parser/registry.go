package parser

// Parser defines the interface for file-level relationship extraction.
type Parser interface {
	CanHandle(ext string) bool
	Parse(absPath string, rootPath string) (*Graph, error)
}

// GoParser implements Parser for Go AST files.
type GoParser struct{}

// CanHandle returns true for Go files.
func (p *GoParser) CanHandle(ext string) bool {
	return ext == ".go"
}

// Parse extracts nodes and edges from Go file.
func (p *GoParser) Parse(absPath string, rootPath string) (*Graph, error) {
	return ParseGoSource(absPath, rootPath)
}

// MarkdownParser implements Parser for markdown debate documents.
type MarkdownParser struct{}

// CanHandle returns true for Markdown files.
func (p *MarkdownParser) CanHandle(ext string) bool {
	return ext == ".md"
}

// Parse extracts nodes and edges from Markdown debate file.
func (p *MarkdownParser) Parse(absPath string, rootPath string) (*Graph, error) {
	return ParseMarkdownDebate(absPath, rootPath)
}

// DefaultParsers indexes the standard enabled language scanners.
var DefaultParsers = []Parser{
	&GoParser{},
	&MarkdownParser{},
}
