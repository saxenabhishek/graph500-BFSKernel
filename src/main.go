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

func construct_graph(filename string, nodes int) {
	outDegree := make([]int, nodes)
	NoOfedges := 0

	c_degree := make(chan Edge)
	go stream_file_on_chan(filename, c_degree)

	for e := range c_degree {
		if e.u == e.v {
			log.Print("ignoring self loop")
			continue
		}
		outDegree[e.u] += 1
		outDegree[e.v] += 1
		NoOfedges += 1
	}

	offsets := make([]int, nodes+1)
	offsets[0] = 0
	for i := 1; i < nodes+1; i++ {
		offsets[i] = outDegree[i-1] + offsets[i-1]
	}
}

func main() {
	const input_file = "input_files/custom_graph.txt"
	const SCALE = 3
	var nodes = int(math.Pow(2, SCALE))
	log.Printf("SCALE is set to %d (%d nodes), reading from file %s", SCALE, nodes, input_file)
	construct_graph(input_file, nodes)
}
