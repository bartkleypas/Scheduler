---
id: visualization-framework
title: Visualization Framework
status: Resolved
decision: Option A (D3.js / Radial Tidy Layout)
targets:
  - web/
---

# Visualization Framework

D3.js was chosen to build a custom, premium "Radial Tidy Tree" projection of the workspace relationships.

## Rationale
The Radial Tidy layout offers a stunningly symmetrical and space-efficient view of workspace hierarchies. D3.js provides raw canvas/SVG coordinate and transition controls to render this layout fluidly while integrating smoothly with static asset embedding in the Go API binary.
