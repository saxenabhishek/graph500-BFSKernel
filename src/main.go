package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Edge struct {
	u int
	v int
}

type CSRGraph struct {
	N         int // vertices
	M         int // undirected edges
	Deg       []int
	Offsets   []int
	Neighbors []int
}

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
		log.Fatalf("Failed to open file \n%s", err)
	}
	defer file.Close()
	defer close(c)
	scn := bufio.NewScanner(file)

	for scn.Scan() {
		line := strings.Fields(scn.Text())
		u, v := line[0], line[1]

		c <- Edge{str2int(u), str2int(v)}
	}

	if err := scn.Err(); err != nil {
		log.Fatalf("Failed to scan file \n%s", err)
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
	offsets[0] = 0
	for i := 1; i < nodes+1; i++ {
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

		start, end := g.Offsets[v], g.Offsets[v+1]
		for i := start; i < end; i++ {
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

func run_bfs_benchmark(g CSRGraph, roots []int) {
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

	for _, root := range roots {
		// time BFS
		t0 := time.Now()
		_, reached = bfs_kernel2(g, root, parent, q, reached)
		dt := time.Since(t0).Seconds()

		times = append(times, dt)
		nedge := count_edges_reached(g, reached)
		// compute metrics

		// if dt is near 0
		tepsVal := 0.0
		if dt > 0 {
			tepsVal = nedge / dt
		}

		teps = append(teps, tepsVal)
		nedges = append(nedges, nedge)

		// reset parent
		for _, v := range reached {
			parent[v] = -1
		}
	}

	stT := stats(times)
	stE := stats(nedges)
	stR := stats(teps)
	hmean := harmonic_mean(teps)

	fmt.Println("scale,nbfs,metric,min,median,max,mean,stddev")

	// BFS time (seconds)
	fmt.Printf("%d,bfs_time_sec,%.6e,%.6e,%.6e,%.6e,%.6e\n",
		NBFS, stT.Min, stT.Median, stT.Max, stT.Mean, stT.Stddev)

	// BFS traversed edges
	fmt.Printf("%d,bfs_nedge,%.0f,%.0f,%.0f,%.0f,%.0f\n",
		NBFS, stE.Min, stE.Median, stE.Max, stE.Mean, stE.Stddev)

	// BFS TEPS
	fmt.Printf("%d,bfs_teps,%.6e,%.6e,%.6e,%.6e,0\n",
		NBFS, stR.Min, stR.Median, stR.Max, hmean)
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

func main() {
	var (
		scale int
		file  string
	)

	// --- 1. Command-line flags ---
	flag.IntVar(&scale, "scale", -1, "Graph scale (N = 2^scale)")
	flag.StringVar(&file, "file", "", "Path to edge list file")
	flag.Parse()

	// --- 2. Environment variables fallback ---
	if scale < 0 {
		if s := os.Getenv("SCALE"); s != "" {
			v, err := strconv.Atoi(s)
			if err != nil {
				log.Fatalf("Invalid SCALE env var: %v", err)
			}
			scale = v
		}
	}

	if file == "" {
		file = os.Getenv("GRAPH_FILE")
	}

	// --- 3. Validate inputs ---
	if scale < 0 {
		log.Fatal("Scale must be provided via --scale or SCALE env var")
	}
	if file == "" {
		log.Fatal("Input file must be provided via --file or GRAPH_FILE env var")
	}

	nodes := 1 << scale

	log.Printf(
		"SCALE=%d (%d nodes), reading from file %s",
		scale, nodes, file,
	)

	log.Printf("SCALE is set to %d (%d nodes), reading from file %s", scale, nodes, file)

	t0 := time.Now()
	g := construct_graph(file, nodes)
	dt := time.Since(t0).Seconds()
	fmt.Printf("construction_time: %20.17e\n", dt)

	roots := sample_roots(g, 64, 49)

	run_bfs_benchmark(g, roots)

}
