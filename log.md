# Change Log

All updates and modifications to this knowledge bundle are tracked chronologically below.

## 2026-07-05

*   **Initialized OKF Bundle**: Created the project-level [index.md](file:///Users/Bart/Code/Scheduler/index.md), [log.md](file:///Users/Bart/Code/Scheduler/log.md), and [graph_schema.md](file:///Users/Bart/Code/Scheduler/graph_schema.md).
*   **Structured Decisions Index**: Added [decisions/index.md](file:///Users/Bart/Code/Scheduler/decisions/index.md) to index all existing Architecture Decision Records.

## 2026-07-06

*   **Optimized Ingestion Performance**: Increased default periodic scan interval to 10 minutes to reduce database noise, and implemented database-backed SHA-256 content hashing (via SQLite migration v3) to completely bypass parsing and database writes for unchanged files.
*   **Added Stale Node and Edge Garbage Collection**: Implemented transactional graph pruning of deleted or renamed files, symbols, debates, decisions, and tasks.
*   **Decoupled Parser Architecture**: Extracted Go AST and Markdown parsers behind a clean, generic `Parser` registry interface to isolate file scanners from the ingestion scheduler.
*   **Refactored Mock Testing**: Unified duplicate test mocks into a centralized `internal/testutil` package, removing test code redundancy.

