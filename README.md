# Graph500 BFS Kernel

Implementation and benchmarking of the Graph500 Breadth-First Search kernel in Go, run on AWS memory-optimized instances to characterize memory-bandwidth-bound graph traversal at scale.

## What is Graph500

Graph500 is an HPC benchmark designed to stress memory subsystems rather than floating-point throughput. It measures **TEPS** (Traversed Edges Per Second) across BFS traversals of large RMAT-generated graphs. Because random graph access patterns defeat hardware prefetchers, performance is bottlenecked by memory bandwidth and latency.

## Implementation

**Graph representation:** CSR (Compressed Sparse Row). TThe RMAT edge list is read twice sequentially, once for degree counting and once for neighbor fill, producing an exponential construction-time curve visible in the results.

**Infrastructure:**
- Two Docker images: one for Kronecker/RMAT graph generation, one for BFS execution
- Logs emitted as JSONL per run (scale, edge factor, TEPS, harmonic mean, runtime, edges traversed)
- Python analytics script generates all plots from the log file

## Environment

`r7` memory-optimized instances chosen because Graph500 is memory-bandwidth-bound. The benefits of multicore scaling are limited at this stage.

## Results Summary

Peak harmonic-mean TEPS at scale 16: **~2.42 × 10⁸** (edge factor 16).

**Cache cliff at scale 20:** TEPS drops sharply between scale 16 and 20 as the working set exceeds LLC capacity and all accesses become DRAM-bound. Beyond scale 20, TEPS stabilizes as the graph sits entirely in memory.

**Edge factor effect:** Edge factor 16 consistently outperforms edge factor 8 in TEPS despite having twice the edges. Denser graphs amortize per-vertex memory latency by providing more useful work per cache line fetch during traversal.

**Variance:** Scale 25 shows noticeably larger error bars across its 64 BFS roots, reflecting sensitivity to starting node selection at that working set size.


## Graph Sizes

ASCII edge list format, (binary would be much smaller and can be a future addition):

| Scale | Edge Factor 8 | Edge Factor 16 |
|---|---|---|
| 16 | 5.9 MB | 12 MB |
| 20 | 111 MB | 223 MB |
| 22 | 495 MB | 991 MB |
| 24 | 2.1 GB | 4.2 GB |
| 26 | 8.9 GB | 18 GB |

## Running

```bash
# Generate graph (scale 22, edge factor 16)
docker run --rm -v $(pwd)/data:/data graph500-generator 22 16

# Run BFS kernel
docker run --rm -v $(pwd)/data:/data -v $(pwd)/logs:/logs graph500-bfs 22 16

# Analyze results
python3 analytics.py logs/results.jsonl
```

## What is not yet done

Parallel BFS across multiple cores/nodes. The sequential kernel provides a memory-bandwidth baseline; parallel results would show where coordination overhead begins to dominate throughput gains. This is the  next step.