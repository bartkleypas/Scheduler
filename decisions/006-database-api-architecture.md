---
id: database-api-architecture
title: Database & API Architecture
status: Resolved
decision: "Pure Go SQLite, UUIDv7 via google/uuid, Formal Migrations, REST/JSON API with Service Decoupling"
targets:
  - pkg/engine/scheduler.go
---

# Database & API Architecture

We resolved the core database engine and API access patterns for Phase 3.

## Rationale
1. **Pure Go SQLite (`modernc.org/sqlite`)**: We chose a Cgo-free SQLite driver to ensure that compiling the scheduler and importing it as a library in other Go codebases requires no C compiler, keeping cross-platform deployments flawless.
2. **UUIDv7 Surrogate Keys (`github.com/google/uuid`)**: UUIDv7 is adopted for surrogate node primary keys. Because they are time-ordered and lexicographically sortable, they prevent SQLite B-tree page fragmentation and enable native, high-speed time-sliced snapshot queries of the knowledge graph.
3. **Formal Migrations**: Using a formal migration framework ensures safe schema updates and B-tree indexing modifications over time as the graph capabilities grow.
4. **Service Interface Pattern**: We decouple the REST/JSON HTTP server from the database querying engine by routing all API calls through a structured `GraphService` interface. This enables REST/JSON today while allowing easy drop-in gRPC adapter integrations tomorrow.
