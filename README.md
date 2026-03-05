# graph500-BFSKernel

The Graph500 benchmark is designed to evaluate data-intensive supercomputer performance using three specific kernels. Kernel 2 (BFS - Breadth-First Search): Performs a parallel Breadth-First Search on the generated graph.

# How to use

## generate graph

Build `dockerfile.kronecker-genrator` image

```bash
 docker run --rm \
      -v ./input_files:/graphs \
      -e SCALE=8 \
      -e EDGE_FACTOR=16 \
      kgraph-gen
```

```bash
go run ./src \
    --scale 8 \
    -ef 16 \
    -file ./input_files/output8.txt \
```

Run docker image

```bash
docker run --rm \
      -v ./input_files:/graphs:ro \
      -v ./output:/app/output \
      -e SCALE=8 \
      -e EDGE_FACTOR=16 \
      -e GRAPH_FILE=/graphs/graph_s8_ef16.txt \
      -e OUT_FILE=/app/output/results.jsonl \
      bfs
```
