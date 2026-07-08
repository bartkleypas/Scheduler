---
type: Decision
title: "Architecture Decision Records (ADRs) Index"
description: "This index tracks key architectural decisions and technical debates throughout the Scheduler lifecycle."
tags:
  - scheduler
  - go
  - decision
  - index
timestamp: "2026-07-05T12:29:48-07:00"
---

# Architecture Decision Records (ADRs) Index

This index tracks key architectural decisions and technical debates throughout the Scheduler lifecycle.

## Decision Records

*   [001-graph-storage](001-graph-storage.md) - SQLite + JSON storage engine selection.
*   [002-visualization-framework](002-visualization-framework.md) - Radial Tidy Layout with D3.js.
*   [003-ingestion-language](003-ingestion-language.md) - Go core with language-agnostic integration hooks.
*   [004-project-layout](004-project-layout.md) - Standard Go layout with reusable `/pkg` package.
*   [005-ingestion-engine](005-ingestion-engine.md) - Event-driven fsnotify and ticker loop.
*   [006-database-api-architecture](006-database-api-architecture.md) - Cgo-free SQLite, UUIDv7 keys, REST API strategy.
*   [007-agent-evaluation-phase3](007-agent-evaluation-phase3.md) - Multi-phase agent integration analysis.
*   [008-agent-evaluation-phase4](008-agent-evaluation-phase4.md) - Assessment of visualizer footprint and agent ergonomics.
*   [009-rejection-of-node-js-visualizer](009-rejection-of-node-js-visualizer.md) - Deletion of Node.js/Vite frontend assets to maintain zero-dependency pure-Go purity.
