---
type: Concept
title: "Project Scheduler Knowledge Index"
description: "This is the project-level entry point for the Scheduler knowledge bundle. Scheduler is a Go-based workspace scanning and indexing engine that build..."
tags:
  - scheduler
  - go
  - index
timestamp: "2026-07-08T19:05:00-07:00"
---

# Project Scheduler Knowledge Index

This is the project-level entry point for the Scheduler knowledge bundle. Scheduler is a Go-based workspace scanning and indexing engine that builds a local-first Knowledge Graph of files, code symbols, technical debates, and decisions.

## Documentation

*   [README](README.md) - Project overview, architecture diagram, setup, and API documentation.

## Subdirectories

*   [decisions](decisions/index.md) - Architecture Decision Records (ADRs) tracking design debates and key trade-offs.

## Concepts

*   [bootstrap_memory](GEMINI.md) - The bootstrap intent document, oriented towards waking agents, detailing active debates and the project roadmap.
*   [graph_schema](graph_schema.md) - Schema modeling nodes (Files, Symbols, Debates, Decisions, Tasks) and edge relationships.

## Research

*   [distributed_build_scheduling](research/distributed_build_scheduling.md) - Distributed build system design, dependency graph scheduling, and optimization trade-offs.
*   [pre_computing_scheduling_evolution](research/pre_computing_scheduling_evolution.md) - Analysis of pre-computing physical scheduling systems and their mapping to modern operating systems.
