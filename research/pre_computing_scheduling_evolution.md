# The Pre-Computing Evolution of Scheduling: From Railroad Sidings to CPU Slices

Before silicon, CPUs, and operating systems, human civilization faced the fundamental challenge of **resource scarcity**. Whether allocating transmission lines, factory floor time, or railway tracks, engineers and operational managers invented scheduling mechanisms to optimize throughput, ensure fairness, and minimize latency.

When digital computing arrived in the mid-20th century, computer scientists did not invent scheduling from scratch. Instead, they formalised and digitized physical, pre-computing practices.

---

## 🗺️ Mapping Physical History to Digital Logic

| Pre-Computing System | Historical Origin | Core Problem Solved | Computer Science Counterpart |
| :--- | :--- | :--- | :--- |
| **Baudot's Telegraph Distributor** (1874) | Émile Baudot (Telegraphy) | Sharing an expensive telegraph wire among multiple operators. | **Round Robin (RR)** & Time-Slicing |
| **Single-Track Sidings** (19th Century) | Railway Dispatchers | Preventing train collisions while letting high-priority trains pass. | **Preemptive Priority Scheduling** |
| **Industrial Manufacturing Shop-Floor** (1950s) | W. E. Smith (Operations Research) | Minimizing the average time raw materials spend waiting to be processed. | **Shortest Job First (SJF)** / Shortest Processing Time |
| **Telephone Trunk Allocation** (1909) | A. K. Erlang (Copenhagen Telephone Exchange) | Sizing switchboards to prevent callers from getting busy signals. | **Queueing Theory** & Network Capacity Planning |
| **Perishable Goods / Postal Deadlines** | Logistics & Agriculture | Prioritizing items before they expire or miss delivery windows. | **Earliest Deadline First (EDF)** |
| **The "Take-A-Number" System** | Delis, Banks, & Bakeries | Ensuring customers are served fairly in the exact order of their arrival. | **First-In, First-Out (FIFO)** / FCFS |

---

## 1. Time-Division Multiplexing (Baudot) $\rightarrow$ Round Robin & Time-Slicing

In the 1870s, the telegraph wire was the highway of information, but it was incredibly expensive to lay. A single operator typing Morse code could not saturate the wire's bandwidth, leading to significant idle time.

```
       Baudot's Rotary Distributor (1874)
   
       Operator 1 ---\       Segmented
       Operator 2 ----\     Time-Slices
       Operator 3 -----*====== Wire ======> Receiver
       Operator 4 ----/ 
       Operator 5 ---/
```

*   **The Physical Method:** French engineer **Émile Baudot** patented a system in 1874 using synchronized rotary distributors (mechanical brushes spinning on segmented wheels). As the brush rotated, it connected each operator's keying device to the line for a fraction of a second in a repeating sequence. Each operator had a brief, repeating slice of time to send a character.
*   **The Digital Adoption:** This mechanical time-division multiplexing (TDM) is the direct ancestor of **Round Robin (RR) CPU scheduling** used in time-sharing operating systems (such as CTSS and Multics in the 1960s). Instead of telegraph operators sharing a wire, multiple active users or threads share a single CPU, with the operating system allocating a millisecond-scale "quantum" (time slice) before rotating to the next.

> [!NOTE]
> The term **"Round Robin"** itself is much older than Baudot. In the 17th century, French and British sailors signing grievances or petitions would write their names in a circle (or on a *ruban rond*, "round ribbon") so that the authorities could not identify the "ringleader" or first signer.

---

## 2. Railway Dispatching $\rightarrow$ Preemptive Priority Scheduling

In 19th-century railway systems, tracks were a scarce and dangerous resource. Running multiple trains on a single track required absolute precision to prevent collisions while keeping transit times low.

```
                  Siding Track (Yield / Blocked)
                     [ Freight Train ]
                    /                 \
    ======== Main Track ============================ Main Track ====>
                             =======>
                       [ Express Passenger ]
```

*   **The Physical Method:** High-priority trains (like express passenger lines carrying mail or VIPs) were given precedence over low-priority trains (like heavy freight). If an express train was running behind, the dispatcher would order the freight train to pull into a **siding** (a short, parallel track). The freight train had to wait there in a "blocked" or "yield" state until the high-priority train roared past on the main line.
*   **The Digital Adoption:** This is the exact physical analog to **Preemptive Priority CPU Scheduling**. In an operating system:
    1. The main track represents the CPU.
    2. The express train represents a high-priority process or hardware interrupt.
    3. The freight train represents a low-priority background process.
    4. The siding represents the operating system placing a thread into a `WAITING` or `READY` state (context switching) to let the urgent task preempt it and run.

---

## 3. Operations Research (Smith's Rule) $\rightarrow$ Shortest Job First (SJF)

In manufacturing plants, a machine shop receives different orders (jobs) of varying sizes. The factory manager's goal is to minimize **mean flow time**—the average time a product spends sitting on the shop floor before completion.

*   **The Physical Method:** In 1956, mathematician and operations researcher **W. E. Smith** published a paper in the *Naval Research Logistics Quarterly* showing that to minimize the average time jobs spend waiting in a single-machine queue, they should be processed in order of their duration—the **Shortest Processing Time (SPT)** rule. If weights are associated with the jobs (representing cost or importance), the jobs should be ordered by their ratio of weight to processing time ($w_j/p_j$), known as Smith's Ratio Rule.
*   **The Digital Adoption:** In computer operating systems, this was adopted as **Shortest Job First (SJF)** or **Shortest Remaining Time First (SRTF)** scheduling. It is mathematically optimal for minimizing average turnaround and waiting time.
    *   *The Catch:* In a factory, you know exactly how long a job will take (based on specifications). In general-purpose computing, the OS cannot predict exactly how long a CPU burst will last before it blocks for I/O, forcing schedulers to predict it using historical averages (exponential smoothing).

---

## 4. Teletraffic Engineering (Erlang) $\rightarrow$ Capacity Planning & Buffers

In 1909, Danish mathematician **Agner Krarup Erlang** worked for the Copenhagen Telephone Company, analyzing how to size the telephone switchboard system.

```
       Callers                      Trunk Lines (Servers)
       [ Caller 1 ] ---\             /--- [ Trunk 1 ]
       [ Caller 2 ] ----*== Queue ==*---- [ Trunk 2 ]
       [ Caller 3 ] ---/             \--- [ Trunk 3 ]
```

*   **The Physical Method:** Erlang needed to determine how many trunk lines were necessary to ensure that callers didn't get busy signals (the **Erlang B** loss formula, where blocked calls are dropped) or wait too long in a queue (the **Erlang C** delay formula, where callers wait for an open operator). 
*   **The Digital Adoption:** Erlang's equations laid the foundations of **Queueing Theory**, which was adopted by computer scientists (like Leonard Kleinrock, who applied it to ARPANET packet switching in the 1960s). It is used today to calculate:
    *   The optimal size of network packet buffers.
    *   The size of thread pools in web servers.
    *   Dynamic scaling algorithms in cloud environments (e.g., Kubernetes HPA).

---

## 5. Logistics & Agriculture $\rightarrow$ Earliest Deadline First (EDF)

*   **The Physical Method:** In postal services and package sorting, letters marked for "Overnight Delivery" are prioritized over standard "Media Mail" because they have an impending deadline. In agriculture and food retail, supermarkets use the **First Expired, First Out (FEFO)** rotation to place milk or produce with the closest expiration dates at the front of the shelf.
*   **The Digital Adoption:** This became **Earliest Deadline First (EDF)**, a dynamic priority scheduling algorithm used in real-time operating systems (RTOS). EDF assigns priorities dynamically based on the deadline of a task: the closer the deadline, the higher the priority. EDF is optimal on preemptive uniprocessors—meaning if a set of tasks can be scheduled without missing deadlines, EDF is guaranteed to find a valid schedule.

---

## 💡 Summary of the Design Philosophy

The history of scheduling shows a clear evolutionary path:

> **Scheduling is not a product of the computer age; it is a mathematical and operational response to resource constraints.** 

When early computer engineers at MIT, DuPont, and Bell Labs were tasked with managing the CPU—the most expensive resource of its time—they naturally looked to the physical systems that had successfully managed telegraph cables, railway tracks, telephone trunk lines, and factory floors for decades.

---

## 🔗 Related Research
* For an analysis of how scheduling principles apply to distributed systems and compilation pipelines, see [Distributed Build Systems & Dependency Graph Scheduling](file:///Users/Bart/Code/Scheduler/research/distributed_build_scheduling.md).

