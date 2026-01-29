package main

import (
	"bufio"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

type Edge struct {
	u int
	v int
}

type CSRGraph struct {
	N         int // vertices
	M         int // undirected edges
	Offsets   []int
	Neighbors []int // length 2*M
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

	log.Println(neighbors)
	log.Println(offsets)

	return CSRGraph{
		N:         nodes,
		M:         NoOfEdges,
		Offsets:   offsets,
		Neighbors: neighbors,
	}
}

func main() {
	const input_file = "input_files/custom_graph.txt"
	const SCALE = 3
	var nodes = int(math.Pow(2, SCALE))
	log.Printf("SCALE is set to %d (%d nodes), reading from file %s", SCALE, nodes, input_file)
	construct_graph(input_file, nodes)
}
