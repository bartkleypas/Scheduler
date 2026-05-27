---
id: ingestion-engine
title: Ingestion Engine & Parser Scope
status: Resolved
decision: "Aspect 1: Unified Event-Driven Loop & Aspect 2: Option A (Go AST + Markdown Focus)"
targets:
  - pkg/engine/scheduler.go
---

# Ingestion Engine & Parser Scope

We resolved the core ingestion detection architecture and set the V1 parsing boundary for the Scheduler.

## Rationale
1. **Aspect 1 (Unified Scheduler Loop)**: Merging filesystem watcher interrupts (`fsnotify`) and wall-clock ticker ticks (`time.Ticker`) into a single non-blocking Go channel pipeline (`chan IngestTrigger`) guarantees eventual consistency, prevents race conditions, and simplifies the Go library import footprint.
2. **Aspect 2 (Go AST & Markdown Scope)**: Restricting V1 parsing exclusively to Go standard library AST parsers and YAML-frontmattered markdown files ensures absolute precision, zero binary weight, and a clean, Cgo-free dependency footprint, making programmatic library imports effortless.
