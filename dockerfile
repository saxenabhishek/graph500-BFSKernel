# Stage 1: build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY src/ ./src/

RUN go build -o bfs_benchmark ./src

# Stage 2: minimal runtime image
FROM alpine:3.19

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bfs_benchmark .

ENV OUT_FILE=/app/output/results.jsonl

# Scale and edge factor — set these at runtime, not baked in
# ENV SCALE=20
# ENV EDGE_FACTOR=16
# ENV GRAPH_FILE=/data/graph.txt

# Create output dir so the binary can always write even without a volume mount
RUN mkdir -p /app/output

ENTRYPOINT ["/app/bfs_benchmark"]
# Pass --scale, --ef, --file as CMD args, or rely on SCALE / EDGE_FACTOR / GRAPH_FILE env vars