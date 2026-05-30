---
id: rejection-of-node-js-visualizer
title: Architectural Purity & Rejection of Node.js Visualizer (Phase 4 Rollback)
status: 🏁 Resolved
decision: "Aborted the Node.js React/D3 visualizer and deleted all web/ and dist/ code to protect the Go scheduler engine from dependency bloat and platform dilution"
targets:
  - pkg/api/server.go
  - GEMINI.md
---

# Architectural Purity & Rejection of Node.js Visualizer (Phase 4 Rollback)

This node documents the architectural decision to abort the Node.js React/D3 visualization layer (originally built in Phase 4) and delete all frontend and asset-embedding code to preserve the absolute purity, lightweight compiling model, and zero-dependency core of the Go scheduler.

## The Architectural Debate
During the initial Phase 4 development, we successfully constructed a high-fidelity React-D3 radial tree visualizer with floating telemetry HUDs, details slide drawers, and monospace terminal streams. While the visualizer looked "almost too good" and worked flawlessly, it introduced a significant architectural compromise:
1. **Dependency Bloat**: It dragged in a massive NPM dependency tree consisting of hundreds of megabytes of nested packages, complex Vite build scripts, and external toolchains.
2. **Platform Dilution**: A project named **Scheduler** represents an event-driven system of absolute control, channel coalescing, B-tree walking, and deterministic concurrency. Grafting a bloated web platform onto the core scheduler diluted the project's soul, shifting the focus from scheduling efficiency to canvas rendering details.
3. **The Philosophy of Node.js**: Just as Ryan Dahl originally introduced Node.js as a focused, asynchronous I/O scheduler, our project needs to double down on what *we* define as the **perfect scheduler** for workspace scanning, rather than serving as another general-purpose web application.

## The Resolution: The Clean "Yoink"
To restore the absolute discipline of the codebase, we decided to completely abort and remove the Node/React web visualizer:
* **Complete Deletion**: The entire `/web` development folder and `/pkg/api/dist` embedded asset bundles were deleted from the repository.
* **REST API Purity**: Modified `pkg/api/server.go` to remove all `//go:embed` directives, `io/fs` imports, and static visualizer routing handlers, reverting the HTTP server back to a high-performance, 100% native Go REST/JSON API server (`GET /api/graph`, `POST /api/ingest`, `DELETE /api/nodes`).
* **Zero Web Footprint**: The project once again compiles instantly, has zero third-party javascript runtimes or package manager overhead, and maintains an extremely clean, lightweight filesystem footprint.

## The Path Forward
By keeping the web visualizer strictly decoupled from the core scheduler:
1. **Headless Modular Library**: The Go scheduler engine remains a clean, importable package (`pkg/engine`, `pkg/parser`, `internal/storage`) that other Go repositories can use to natively schedule background scanning and AST indexing.
2. **Decoupled API Contract**: The server continues to serve clean JSON representations of the Knowledge Graph. Developers who wish to visualize the workspace can easily pipe this API data into lightweight, separate viewer utilities without imposing any dependency bloat on the scheduler binary itself.
