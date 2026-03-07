#!/usr/bin/env bash
# 1. Builds both Docker images
# 2. Generates all (scale, edge_factor) graphs into ./graphs/
# 3. Runs the BFS benchmark for every config, appending to ./output/results.jsonl

set -euo pipefail


Customise the SCALES and EDGE_FACTORS arrays below.
SCALES=(23)
EDGE_FACTORS=(8 16)

GRAPHS_DIR="$(pwd)/graphs"
OUTPUT_DIR="$(pwd)/output"

GENERATOR_IMAGE="graph500-generator"
BFS_IMAGE="graph500-bfs"

mkdir -p "$GRAPHS_DIR" "$OUTPUT_DIR"

echo "════════════════════════════════════════"
echo " Graph500 Sweep Runner"
echo " Scales:       ${SCALES[*]}"
echo " Edge factors: ${EDGE_FACTORS[*]}"
echo " Graphs dir:   $GRAPHS_DIR"
echo " Output dir:   $OUTPUT_DIR"
echo "════════════════════════════════════════"

echo ""
echo "── Building generator image ──"
docker build -f dockerfile.kronecker-genrator -t "$GENERATOR_IMAGE" .

echo ""
echo "── Building BFS benchmark image ──"
docker build -f Dockerfile -t "$BFS_IMAGE" .

echo ""
echo "── Generating graphs ──"

for SCALE in "${SCALES[@]}"; do
  for EF in "${EDGE_FACTORS[@]}"; do
    GRAPH_FILE="$GRAPHS_DIR/graph_s${SCALE}_ef${EF}.txt"

    if [[ -f "$GRAPH_FILE" ]]; then
      echo "  [skip] graph_s${SCALE}_ef${EF}.txt already exists"
      continue
    fi

    echo "  Generating scale=${SCALE} ef=${EF} ..."
    docker run --rm \
      -v "$GRAPHS_DIR":/graphs \
      -e SCALE="$SCALE" \
      -e EDGE_FACTOR="$EF" \
      "$GENERATOR_IMAGE"

    echo "  Done -> $(du -sh "$GRAPH_FILE" | cut -f1)"
  done
done

echo ""
echo "── All graphs ready ──"
ls -lh "$GRAPHS_DIR"

echo ""
echo "── Running BFS benchmarks ──"

for SCALE in "${SCALES[@]}"; do
  for EF in "${EDGE_FACTORS[@]}"; do
    GRAPH_FILE="/graphs/graph_s${SCALE}_ef${EF}.txt"

    echo "  Benchmarking scale=${SCALE} ef=${EF} ..."
    docker run --rm \
      -v "$GRAPHS_DIR":/graphs:ro \
      -v "$OUTPUT_DIR":/app/output \
      -e SCALE="$SCALE" \
      -e EDGE_FACTOR="$EF" \
      -e GRAPH_FILE="$GRAPH_FILE" \
      -e OUT_FILE=/app/output/results.jsonl \
      "$BFS_IMAGE"
  done
done

echo ""
echo "════════════════════════════════════════"
echo " All runs complete."
echo " Results: $OUTPUT_DIR/results.jsonl"
echo "════════════════════════════════════════"

cd analytics/
uv run main.py --input /Users/as712/Projects/graph500-BFSKernel/output/results.jsonl