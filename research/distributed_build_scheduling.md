# Distributed Build Systems & Dependency Graph Scheduling

Modern build systems and distributed compilers (such as `distcc`, `icecc`, `sccache`, and `Bazel`) solve the problem of compiling massive codebases by scheduling tasks across a network of machines. This domain sits at the intersection of graph theory, distributed systems, and queueing theory.

---

## 1. The Core Data Structure: The Dependency DAG

A build system represents a software project as a **Directed Acyclic Graph (DAG)** of files and build actions.

```
       [ header.h ] 
       /          \
  [ main.cpp ]  [ helper.cpp ]
       \          /
     [ compiler task ]
       /          \
   [ main.o ]   [ helper.o ]
       \          /
       [ linker task ]
              |
       [ executable ]
```

### Graph Components
* **Nodes**: Represent files (source, headers, object files, libraries) and actions (compiling, linking, code generation).
* **Edges**: Represent dependencies. An edge from node $A$ to node $B$ indicates that $B$ must be built before $A$ can be processed.

### Scheduling Implications
A build scheduler traverses this DAG to plan execution:
* **Concurrency**: Independent subgraphs (e.g., compiling `main.cpp` and `helper.cpp`) can be scheduled concurrently on different processor cores or network nodes.
* **Synchronization**: Downstream tasks (e.g., the linker task) act as synchronization barriers, blocking until all parent nodes are resolved.
* **Path Analysis**: The scheduler identifies the **critical path**—the longest sequence of dependent tasks. The total build time is bounded by the time required to execute this path.

---

## 2. Pre-Computing Physical Analogs

The concepts underlying distributed compilation and graph scheduling have direct historical precedents in manual, physical processes.

### The Medieval Scriptorium (Parallel Copying & Assembly)
Before the printing press, reproducing a long manuscript was a slow, sequential task for a single scribe. To accelerate the process, scriptoriums adopted a distributed workflow:
* **Decomposition (Parsing)**: The head scribe (the scheduler) unbound the original codex into individual sections called **quires** (typically groups of 4 leaves or 8 pages).
* **Distribution (Scheduling)**: The quires were distributed to multiple scribes (worker nodes) working in parallel in the scriptorium.
* **Linking (Binding)**: Once all scribes completed their respective quires, the bookbinder gathered the completed sections, verified their order, and bound them into a single volume (the executable).
* **Dependency Constraint**: If scribe 3 fell ill, the bookbinder could not complete the volume, even if all other scribes finished early. Scribe 3 was on the critical path.

### Industrial Subcontracting (Supply Chain DAGs)
In heavy manufacturing, such as 20th-century shipbuilding, the construction of a vessel is modeled as a complex assembly graph:
* **Local vs. Remote Allocation**: The primary shipyard (the orchestrator) fabricates the hull locally but subcontracts components (engines, propellers, electrical systems) to external factories (remote workers).
* **Interface Contracts**: Subcontractors must build components to precise specifications (similar to compiler API contracts or header files) to ensure they fit during assembly.
* **Assembly (Linking)**: The components are shipped back to the dry dock for final integration. A delay in receiving the propeller blocks the final assembly, leaving the shipyard idle.

---

## 3. The Distributed Scheduling Problem

When a compiler scheduler decides whether to delegate a task to a remote machine, it must evaluate a multi-processor scheduling problem with communication delays.

### The Trade-off Equation
To achieve a net speedup, the cost of distributing a job must be lower than running it locally:

$$T_{\text{distributed}} < T_{\text{local}}$$

Where:
* $T_{\text{local}} = T_{\text{compute\_local}} + T_{\text{queue\_local}}$
* $T_{\text{distributed}} = T_{\text{serialize}} + T_{\text{transmit\_input}} + T_{\text{queue\_remote}} + T_{\text{compute\_remote}} + T_{\text{transmit\_output}} + T_{\text{deserialize}}$

### Decision Factors
1. **Network Bandwidth & Latency**: Sending large source files or deep header dependency trees over a slow network can make remote execution slower than local execution.
2. **Task Granularity**: Compiling a small file with few dependencies is often kept local. Compiling a large file with heavy template instantiations is a better candidate for remote scheduling.
3. **Worker Capacity**: Schedulers must track the CPU core count, clock speed, and current queue depth of each network worker to avoid overloading a single machine.

---

## 4. Mitigation & Optimization Strategies

To handle the unreliability of networks and remote hardware, distributed build schedulers use specific optimization techniques.

### Content-Addressable Caching (Memoization)
To avoid scheduling unnecessary work, modern systems (like `sccache` or Bazel's shared cache) employ content-addressed storage:
* The system hashes the exact contents of all input files, compiler options, env variables, and the compiler toolchain binary.
* This hash serves as a unique key.
* If a remote worker or local cache already contains an entry for this key, the compiler skips scheduling and compilation entirely, downloading the cached output directly.

### Speculative Execution (Straggler Mitigation)
In large-scale distributed systems, a small percentage of worker nodes will experience transient slowdowns (due to thermal throttling, disk I/O spikes, or network congestion). These nodes are called **stragglers**.
* If a single compilation unit on the critical path is scheduled on a straggler, the entire build is delayed.
* **Mitigation**: The scheduler monitors execution times. If a task takes significantly longer than expected, the scheduler spawns a duplicate of the same task on a different worker.
* Whichever worker completes the task first writes to the build target, and the scheduler cancels the outstanding duplicate task.

---

## 5. Summary of System Classifications

| Tool | Graph Definition | Scheduling Mechanism | Caching Strategy |
| :--- | :--- | :--- | :--- |
| **`distcc`** | Defined by local `make` or `ninja`. | Preprocesses source locally; distributes flat compilation tasks to remote daemons. | None (relies on local `ccache` integration). |
| **`sccache`** | Defined by build tool. | Local compilation wrapper; intercepts calls to distribute or cache. | Global storage (S3, Redis, Memcached) matching inputs hash. |
| **`Bazel`** | Explicitly defined in `BUILD` files (Action Graph). | Orchestrates a strict, hermetic DAG. Schedules actions locally or via Remote Execution API. | Strict content-addressed cache of inputs, outputs, and action definitions. |
