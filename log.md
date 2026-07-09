---
type: Concept
title: "Change Log"
description: "All updates and modifications to this knowledge bundle are tracked chronologically below."
tags:
  - scheduler
  - go
  - log
timestamp: "2026-07-09T14:30:00-07:00"
---

# Change Log

All updates and modifications to this knowledge bundle are tracked chronologically below.

## 2026-07-05

*   **Initialized OKF Bundle**: Created the project-level [index.md](index.md), [log.md](log.md), and [graph_schema.md](graph_schema.md).
*   **Structured Decisions Index**: Added [decisions/index.md](decisions/index.md) to index all existing Architecture Decision Records.

## 2026-07-06

*   **Optimized Ingestion Performance**: Increased default periodic scan interval to 10 minutes to reduce database noise, and implemented database-backed SHA-256 content hashing (via SQLite migration v3) to completely bypass parsing and database writes for unchanged files.
*   **Added Stale Node and Edge Garbage Collection**: Implemented transactional graph pruning of deleted or renamed files, symbols, debates, decisions, and tasks.
*   **Decoupled Parser Architecture**: Extracted Go AST and Markdown parsers behind a clean, generic `Parser` registry interface to isolate file scanners from the ingestion scheduler.
*   **Refactored Mock Testing**: Unified duplicate test mocks into a centralized `internal/testutil` package, removing test code redundancy.

## 2026-07-07

*   **Reduced Log Verbosity**: Introduced a `Verbose` configuration option to `Config` to suppress high-volume log outputs by default, including the full knowledge graph database snapshot, unsupported file extension warnings, and full workspace scan progress notifications.
*   **Fixed Absolute path Resolution**: Resolved relative configuration workspace paths (e.g. `"."`) to absolute paths during scheduler instantiation to prevent `filepath.Rel` failures and erroneous database pruning of file nodes.

## 2026-07-08

*   **Indexed Research Documents**: Registered research documents in the project knowledge index ([index.md](index.md)), covering [distributed_build_scheduling.md](research/distributed_build_scheduling.md) and [pre_computing_scheduling_evolution.md](research/pre_computing_scheduling_evolution.md).

## 2026-07-09

*   **Created Project README**: Documented the Scheduler architecture, key features, database schema mapping, project layout, CLI setup commands, API routes, and developer orientation in [README.md](README.md).
*   **Linked README in Knowledge Index**: Registered the new readme document in the project knowledge index ([index.md](index.md)).
*   **Streamlined Agent Bootstrap Memory**: Refactored [GEMINI.md](GEMINI.md) into a thin orientation anchor, deleting redundant schemas and resolved debates, and converting it to a link-based hub pointing to authoritative source files.
