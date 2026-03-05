package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func str2int(vtx string) int {
	u_i, err := strconv.Atoi(vtx)
	if err != nil {
		log.Fatal("Invalid value in file")
	}
	return u_i
}

func stream_file_on_chan(filename string, c chan Edge) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file\n%s", err)
	}
	defer file.Close()
	defer close(c)
	scn := bufio.NewScanner(file)

	for scn.Scan() {
		line := strings.Fields(scn.Text())
		c <- Edge{str2int(line[0]), str2int(line[1])}
	}
	if err := scn.Err(); err != nil {
		log.Fatalf("Failed to scan file\n%s", err)
	}
}

func construct_graph(filename string, nodes int) CSRGraph {
	outDegree := make([]int, nodes)
	NoOfEdges := 0

	c_degree := make(chan Edge)
	go stream_file_on_chan(filename, c_degree)

	for e := range c_degree {
		if e.u == e.v {
			continue
		}
		outDegree[e.u]++
		outDegree[e.v]++
		NoOfEdges++
	}

	offsets := make([]int, nodes+1)
	for i := 1; i <= nodes; i++ {
		offsets[i] = outDegree[i-1] + offsets[i-1]
	}

	// size of length twice of all edges
	neighbors := make([]int, offsets[nodes])

	// temp copy for use in next step
	next := make([]int, nodes)
	copy(next, offsets[:nodes])

	c_Edges := make(chan Edge)
	go stream_file_on_chan(filename, c_Edges)

	for e := range c_Edges {
		if e.u == e.v {
			continue
		}

		neighbors[next[e.u]] = e.v
		next[e.u]++

		neighbors[next[e.v]] = e.u
		next[e.v]++
	}

	return CSRGraph{
		N:         nodes,
		M:         NoOfEdges,
		Deg:       outDegree,
		Offsets:   offsets,
		Neighbors: neighbors,
	}
}

func bfs_kernel2(g CSRGraph, root int, parent []int, q []int, reached []int) ([]int, []int) {
	reached = reached[:0]

	parent[root] = root
	reached = append(reached, root)

	head, tail := 0, 0
	q[tail] = root
	tail++

	for head < tail {
		v := q[head]
		head++

		for i := g.Offsets[v]; i < g.Offsets[v+1]; i++ {
			u := g.Neighbors[i]
			if parent[u] == -1 {
				parent[u] = v
				reached = append(reached, u)
				q[tail] = u
				tail++
			}
		}
	}
	return parent, reached
}

func count_edges_reached(g CSRGraph, reached []int) float64 {
	sum := 0
	for _, v := range reached {
		sum += g.Deg[v]
	}
	return float64(sum) / 2.0
}

func run_bfs_benchmark(
	g CSRGraph,
	roots []int,
	scale int,
	edgeFactor int,
	constructTime float64,
	writer *bufio.Writer,
) {
	NBFS := len(roots)

	parent := make([]int, g.N)
	for i := range parent {
		parent[i] = -1
	}
	q := make([]int, g.N)
	reached := make([]int, 0, g.N)

	times := make([]float64, 0, NBFS)
	nedges := make([]float64, 0, NBFS)
	teps := make([]float64, 0, NBFS)

	enc := json.NewEncoder(writer)

	for i, root := range roots {
		t0 := time.Now()
		_, reached = bfs_kernel2(g, root, parent, q, reached)
		dt := time.Since(t0).Seconds()

		nedge := count_edges_reached(g, reached)
		tepsVal := 0.0
		if dt > 0 {
			tepsVal = nedge / dt
		}

		times = append(times, dt)
		nedges = append(nedges, nedge)
		teps = append(teps, tepsVal)

		// Write one run record
		rec := RunRecord{
			Type:       "run",
			Scale:      scale,
			EdgeFactor: edgeFactor,
			RunIndex:   i,
			Root:       root,
			TimeS:      dt,
			Nedge:      nedge,
			TEPS:       tepsVal,
		}
		if err := enc.Encode(rec); err != nil {
			log.Printf("Warning: failed to write run record: %v", err)
		}

		// Reset parent array
		for _, v := range reached {
			parent[v] = -1
		}
	}

	// Write summary record
	stT := stats(times)
	stE := stats(nedges)
	stR := stats(teps)

	summary := SummaryRecord{
		Type:           "summary",
		Scale:          scale,
		EdgeFactor:     edgeFactor,
		Nodes:          g.N,
		TotalEdges:     g.M,
		NBFS:           NBFS,
		ConstructTimeS: constructTime,

		TimeMin:    stT.Min,
		TimeMedian: stT.Median,
		TimeMax:    stT.Max,
		TimeMean:   stT.Mean,
		TimeStddev: stT.Stddev,

		NedgeMin:    stE.Min,
		NedgeMedian: stE.Median,
		NedgeMax:    stE.Max,
		NedgeMean:   stE.Mean,
		NedgeStddev: stE.Stddev,

		TEPSMin:      stR.Min,
		TEPSMedian:   stR.Median,
		TEPSMax:      stR.Max,
		TEPSHarmonic: harmonic_mean(teps),
	}

	if err := enc.Encode(summary); err != nil {
		log.Printf("Warning: failed to write summary record: %v", err)
	}

	writer.Flush()

	// print summary to stdout for visibility
	fmt.Printf("[DONE] scale=%d ef=%d  harmonic_TEPS=%.4e  median_time=%.4es\n",
		scale, edgeFactor, summary.TEPSHarmonic, stT.Median)
}

func sample_roots(g CSRGraph, NBFS int, seed int64) []int {
	r := rand.New(rand.NewSource(seed))
	cands := make([]int, 0, g.N)

	for v := 0; v < g.N; v++ {
		if g.Deg[v] > 0 {
			cands = append(cands, v)
		}
	}

	r.Shuffle(len(cands), func(i, j int) { cands[i], cands[j] = cands[j], cands[i] })
	if len(cands) > NBFS {
		cands = cands[:NBFS]
	}
	return cands
}

func getConfig() (int, int, string, string) {
	var (
		scale      int
		edgeFactor int
		file       string
		outFile    string
	)

	flag.IntVar(&scale, "scale", -1, "Graph scale (N = 2^scale)")
	flag.IntVar(&edgeFactor, "ef", 16, "Average edges per vertex (edge factor)")
	flag.StringVar(&file, "file", "", "Path to edge list file")
	flag.StringVar(&outFile, "out", "", "Path to JSONL output file (overrides OUT_FILE env)")
	flag.Parse()

	// Environment variable fallbacks
	if scale < 0 {
		if s := os.Getenv("SCALE"); s != "" {
			v, err := strconv.Atoi(s)
			if err != nil {
				log.Fatalf("Invalid SCALE env var: %v", err)
			}
			scale = v
		}
	}
	if edgeFactor == 16 {
		if s := os.Getenv("EDGE_FACTOR"); s != "" {
			v, err := strconv.Atoi(s)
			if err != nil {
				log.Fatalf("Invalid EDGE_FACTOR env var: %v", err)
			}
			edgeFactor = v
		}
	}
	if file == "" {
		file = os.Getenv("GRAPH_FILE")
	}
	if outFile == "" {
		outFile = os.Getenv("OUT_FILE")
	}
	if outFile == "" {
		outFile = "output/results.jsonl"
	}

	// Validate
	if scale < 0 {
		log.Fatal("Scale must be provided via --scale or SCALE env var")
	}
	if file == "" {
		log.Fatal("Input file must be provided via --file or GRAPH_FILE env var")
	}
	return scale, edgeFactor, file, outFile
}

func main() {
	scale, edgeFactor, file, outFile := getConfig()

	nodes := 1 << scale
	log.Printf("SCALE=%d (%d nodes), EF=%d, reading from %s", scale, nodes, edgeFactor, file)
	log.Printf("Output will be written to: %s", outFile)

	// Create output directory if needed
	if dir := filepath.Dir(outFile); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}
	}

	f, err := os.OpenFile(outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open output file: %v", err)
	}
	defer f.Close()
	writer := bufio.NewWriter(f)

	// Construct graph
	t0 := time.Now()
	g := construct_graph(file, nodes)
	constructTime := time.Since(t0).Seconds()
	log.Printf("Graph constructed in %.4fs  (N=%d, M=%d)", constructTime, g.N, g.M)

	roots := sample_roots(g, 64, 49)
	run_bfs_benchmark(g, roots, scale, edgeFactor, constructTime, writer)
}
