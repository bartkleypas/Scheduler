---
id: ingestion-language
title: Ingestion Language & Integration Hooks
status: Resolved
decision: Option A (Go-Based Core with Language-Agnostic Hooks)
targets:
  - pkg/engine/scheduler.go
---

# Ingestion Language & Integration Hooks

Go was selected as the backend parser and scheduler core.

## Rationale
Go compiles into a single standalone binary with zero external execution dependencies. To enable reusability across workspace repositories, the engine is structured as an importable Go package library, while exposing standard JSON-over-stdin CLI hooks so other languages can pipe metadata easily.
