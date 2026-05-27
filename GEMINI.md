# Project Scheduler: Workspace Knowledge Graph & Decision Tracker

> **Attention Agents / Context Windows:** This file serves as the system "bootstrap" memory and the seed of the local Workspace Knowledge Graph. When you wake up in this workspace, read this file to understand the current state, project intent, ongoing architectural debates, and decision history.

---

## 🌌 Project Vision

**Scheduler** is a local-first system designed to model, build, and maintain a **Knowledge Graph of the active workspace**. It captures not just code symbols and files, but the *intent, architecture decision records (ADRs), trade-offs, and design debates* that shape the project.

The "Scheduler" component refers to the automated background engine that:
1. Schedules periodic or change-triggered scans of the workspace.
2. Parses code syntax, markdown debates, and tasks.
3. Indexes and updates the Knowledge Graph in real-time.
4. Feeds an interactive, highly visual 2D/3D force-directed dashboard.

---

## 🧠 Knowledge Graph Schema Concept

The graph models the codebase through standard Node and Edge relationships:

### Nodes
- `File`: A source or markdown file.
- `Symbol`: Functions, classes, components, or modules.
- `Debate`: An ongoing technical discussion or unresolved design challenge.
- `Option`: A proposed solution/alternative within a Debate.
- `Decision`: A resolved Debate with associated rationale.
- `Task`: A concrete action item linked to files or decisions.

### Edges
- `(File) -[IMPORTS]-> (File)`
- `(Symbol) -[DEFINED_IN]-> (File)`
- `(Debate) -[TARGETS]-> (File | Symbol)`
- `(Debate) -[PROPOSES]-> (Option)`
- `(Decision) -[APPROVED]-> (Option)`
- `(Task) -[RESOLVES]-> (Debate)`

---

## 💬 Active Debates & Architectural Decisions

This section captures the ongoing debates regarding the system's design. As developers (human or AI) propose or decide on paths, they should update the status of these blocks.

### ⚖️ Debate #1: Graph Storage Engine
* **Status**: 🏁 Resolved
* **Decision**: **Option A (SQLite + JSON)**
* **Rationale**: SQLite provides a lightweight, highly queryable SQL engine perfectly suited for local relational schemas (nodes and edges), while JSON columns or exports guarantee compatibility, ease of synchronization, and version-controlled snapshots.

### ⚖️ Debate #2: Visualization Framework
* **Status**: 🏁 Resolved
* **Decision**: **Option A (D3.js / Radial Tidy Layout)**
* **Rationale**: D3.js provides the exact high-precision, low-level coordinate manipulation (polar coordinates, custom curve projections) required to build a gorgeous, mathematically perfect **Radial Tidy Tree** graph visualization of the workspace. Its buttery smooth transitions, dynamic scaling, and rich styling hook into Go's embedded HTTP server perfectly, creating a highly premium user interface.

### ⚖️ Debate #3: Ingestion Language & Integration Hooks
* **Status**: 🏁 Resolved
* **Decision**: **Option A (Go-Based Core with Language-Agnostic Hooks)**
* **Rationale**: Go is selected as the primary language for the core scheduler and parser engine. To make the project extremely flexible:
  - **Go Importable Package**: The engine is structured as a modular, importable Go package. This allows other Go repositories to directly import it as a library to natively integrate graph-scheduling capabilities.
  - **Language-Agnostic CLI Hooks**: The scheduler exposes simple JSON/stdin CLI integration hooks, allowing non-Go codebases (JS, Python, etc.) to pipe workspace metadata into the graph without needing custom Go environments.

### ⚖️ Debate #4: Project Layout & Module Architecture
* **Status**: 🏁 Resolved
* **Decision**: **Option A (Standard Go Layout with Reusable `/pkg` & YAML Frontmatter Debates)**
* **Rationale**: Adopted the standard Go project structure with a reusable public `/pkg` to allow other Go repositories to import the scheduler library, keeping internal details protected in `/internal`. Configured Markdown debates to use structured YAML Frontmatter to guarantee highly reliable, programmatic indexing of design decisions and edge relationships. To make the project extremely flexible, we decoupled the parsing machinery into its own dedicated public package: `pkg/parser`. This allows other Go codebases to import the AST/Markdown parser as a standalone library without needing the scheduling or watcher runtimes.

### ⚖️ Debate #5: Ingestion Engine & Parser Scope (Phase 2 Start)
* **Status**: 🏁 Resolved
* **Decision**: **Aspect 1: Unified Event-Driven Loop** & **Aspect 2: Option A (Go AST + Markdown Focus)**
* **Rationale**: 
  - **Aspect 1**: We will merge FS watcher interrupts (`fsnotify`) and wall-clock tickers into a unified non-blocking `chan IngestTrigger` channel pipeline in Go. Highly performant, race-free, and flexible.
  - **Aspect 2**: Target exclusively Go AST files and YAML-frontmattered markdown files for V1. This avoids complex Cgo dependencies, provides total structural precision, and ensures the library is lightweight and clean to import into other Go repositories.

### ⚖️ Debate #6: Database & API Architecture (Phase 3 Start)
* **Status**: 🏁 Resolved
* **Decision**: 
  - **SQLite Driver**: `modernc.org/sqlite` (pure Go, Cgo-free).
  - **Surrogate Keys**: UUIDv7 via `github.com/google/uuid` (lexicographically sortable for time-sliced queries).
  - **Migrations**: Formal migrations library for long-term schema safety.
  - **API Strategy**: RESTful HTTP + JSON API backed by a decoupled `GraphService` interface to easily plug in gRPC later.
* **Rationale**: Approved by the user. Pure Go SQLite keeps cross-compilation flawless. UUIDv7 provides monotonically increasing IDs, preventing index fragmentation and enabling temporal queries. The Service Interface pattern separates the core query engine from transport layers, allowing REST/JSON today and seamless gRPC integration tomorrow.

---


## 🗺️ Project Roadmap

- [x] **Phase 1: Knowledge Seed & Design**
  - [x] Establish the `GEMINI.md` bootstrap memory file.
  - [x] Agree on debate formats and file/folder structures.
- [x] **Phase 2: The Ingestion Engine (Backend CLI)**
  - [x] Implement local watcher / scheduler.
  - [x] Write the parser for markdown debates and basic Go AST parsing.
- [ ] **Phase 3: Database & API (Current)**
  - [ ] Define the graph storage schema.
  - [ ] Create simple local query API.
- [ ] **Phase 4: Stunning Visualizer Dashboard**
  - [ ] Build the interactive network graph interface.
  - [ ] Design the glassmorphic controls and live ingest logs.

---

## 📟 Instructions for Waking Agents

If you are a newly instantiated agent in this workspace:
1. **Always read this file first** to orient yourself.
2. **Contribute to debates** by adding options, refining pros/cons, or moving debates to `Status: Resolved` (under a new `Decision` heading) once the user approves.
3. **Keep this file updated** as the single source of truth for design decisions and roadmap progress. Do not allow it to fall out of sync with actual file architectures.
