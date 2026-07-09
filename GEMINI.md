---
type: Concept
title: "Project Scheduler: Agent Bootstrap Memory"
description: "Immediate orientation anchor for developers and AI agents waking up in the Scheduler project."
tags:
  - scheduler
  - go
  - bootstrap
timestamp: "2026-07-09T14:30:00-07:00"
---

# Project Scheduler: Agent Bootstrap Memory

> [!IMPORTANT]
> **Attention Agents / Context Windows:** This file serves as the system "bootstrap" memory. Read this first to understand the *immediate* state of the project, open debates, and active tasks. 

---

## 🗺️ Project Navigation Hub

For detailed and authoritative documentation, consult these dedicated files instead of relying on this bootstrap anchor:
*   [Project Index](index.md) - The central directory of concepts, decisions, and research.
*   [Technical README](README.md) - Build instructions, architecture diagram, design details, and API documentation.
*   [Graph Schema](graph_schema.md) - Node and edge model definitions.
*   [Architecture Decisions (ADRs)](decisions/index.md) - Historical record of all resolved design debates.
*   [Change Log](log.md) - Chronological ledger of all modifications.

---

## 🏗️ Active Design Context & Focus

*   **Current Focus**: Stabilizing the core Go API-only engine and scheduling loops after rolling back the React visualization footprint (per [ADR #009](decisions/009-rejection-of-node-js-visualizer.md)).
*   **Next Steps**: Preparing the package for library imports by other Go projects in the workspace.

---

## ⚖️ Active Debates

*   *There are currently no active debates.* All historical debates have been resolved and moved to [decisions/index.md](decisions/index.md).

---

## 🗺️ Next Up (Roadmap)

- [ ] **Phase 5: Importable Library Integration**
  - [ ] Verify external import paths.
  - [ ] Write integration examples for downstream workspaces (e.g. `please`).
