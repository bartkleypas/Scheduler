---
type: Concept
title: Knowledge Graph Schema
description: Schema modeling nodes and edge relationships for representing code syntax, directories, debates, options, and decisions.
tags:
- schema
- knowledge-graph
- databases
- sqlite
timestamp: '2026-07-05T12:29:00-07:00'
---

# Knowledge Graph Schema

The Scheduler engine constructs a relational graph database (`graph.db`) modeling files, programming symbols, and the decisions surrounding them.

## Graph Model

### Nodes

*   `File`: A source code file (e.g., `.go`, `.py`, `.js`) or a Markdown documentation file.
*   `Symbol`: Concrete programming entities (functions, structs, classes, variables, modules).
*   `Debate`: An active design decision or unresolved engineering challenge.
*   `Option`: A proposed path or solution to an open Debate.
*   `Decision`: A resolved Debate detailing the selected Option and its rationale.
*   `Task`: A concrete TODO item linked to a file or debate resolution.

---

### Edges (Relationships)

| Source Node | Edge Label | Target Node | Purpose |
| --- | --- | --- | --- |
| `File` | `IMPORTS` | `File` | Tracks import graphs/dependencies. |
| `Symbol` | `DEFINED_IN` | `File` | Tracks location definitions. |
| `Debate` | `TARGETS` | `File` \| `Symbol` | Identifies components affected by design debates. |
| `Debate` | `PROPOSES` | `Option` | Lists options proposed as solutions. |
| `Decision` | `APPROVED` | `Option` | Logs which option resolved a debate. |
| `Task` | `RESOLVES` | `Debate` | Links outstanding tasks to the debates they clear. |
