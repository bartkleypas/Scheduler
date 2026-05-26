---
id: graph-storage
title: Graph Storage Engine
status: Resolved
decision: Option A (SQLite + JSON)
targets:
  - pkg/engine/scheduler.go
---

# Graph Storage Engine

SQLite was chosen as the relational local storage layer for the knowledge graph. 

## Rationale
SQLite provides lightweight, zero-dependency, and extremely fast local database storage. Relationships between code symbols, files, and markdown debates (nodes and edges) are stored in clean SQLite tables, while rich properties are stored in a flexible JSON format inside columns to ensure seamless export/sync capabilities.
