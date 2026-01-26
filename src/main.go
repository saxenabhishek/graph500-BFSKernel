package main

import (
	"bufio"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

func inc_degree_count(outDegree []int, vtx string) {
	u_i, err := strconv.Atoi(vtx)
	if err != nil {
		log.Fatal("Invalid value in file")
	}
	outDegree[u_i] += 1
}

func stream_file_on_chan(filename string, c chan string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file \n%s", err)
	}
	defer file.Close()
	defer close(c)
	scn := bufio.NewScanner(file)

	for scn.Scan() {
		line := scn.Text()
		c <- line
	}

	if err := scn.Err(); err != nil {
		log.Fatalf("Failed to scan file \n%s", err)
	}

}

func construct_graph(filename string, nodes int) {
	outDegree := make([]int, nodes)
	edges := 0

	c := make(chan string)
	go stream_file_on_chan(filename, c)

	for line := range c {
		edge := strings.Fields(line)
		u, v := edge[0], edge[1]
		if u == v {
			log.Print("ignoring self loop")
			continue
		}
		inc_degree_count(outDegree, u)
		inc_degree_count(outDegree, v)
		edges += 1
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
