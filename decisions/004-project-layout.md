---
id: project-layout
title: Project Layout & Module Architecture
status: Resolved
decision: Option A (Standard Go Layout with Reusable /pkg & YAML Frontmatter Debates)
targets:
  - cmd/
  - pkg/
  - internal/
  - decisions/
---

# Project Layout & Module Architecture

We structured the repository around standard Go design layout schemas, while defining YAML Frontmatter as the standardized debate parser format.

## Rationale
Decoupling public `/pkg` libraries from internal machinery ensures safe imports. Frontmatter blocks allow the file system watcher and AST parser to extract graph nodes and links programmatically with complete precision.
