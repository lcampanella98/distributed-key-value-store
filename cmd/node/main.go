package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/lcampanella98/distributed-key-value-store/internal/api"
	"github.com/lcampanella98/distributed-key-value-store/internal/benchmarks"
	"github.com/lcampanella98/distributed-key-value-store/internal/client"
	"github.com/lcampanella98/distributed-key-value-store/internal/cluster"
)

func main() {
	nodesPtr := flag.String("nodes", "", "pass a comma separated list of node hostnames")
	portPtr := flag.Int("port", 0, "pass a port for this node to run on")
	replicasPtr := flag.Int("replicas", 0, "the number of total nodes including the owner on which each piece of data lives")
	writeModePtr := flag.String("writeMode", "", "strict or best_effort")

	flag.Parse()

	nodeNames := strings.Split(*nodesPtr, ",")
	port := *portPtr
	replicas := *replicasPtr
	writeMode := *writeModePtr

	if replicas < 1 {
		panic("Replicas must be at least 1. Remember replicas includes the primary")
	}
	if replicas > len(nodeNames) {
		panic("Replicas must not exceed the number of total nodes")
	}

	cluster.SetCoordinatorConfig(writeMode)

	var nodes []cluster.Node
	thisNodeIdx := -1
	for i, name := range nodeNames {
		node := cluster.Node{Name: name, Addr: "http://" + name}
		nodes = append(nodes, node)
		if strings.Contains(name, strconv.Itoa(port)) {
			fmt.Printf("Found this node as %s\n", node.Name)
			thisNodeIdx = i
		}
	}
	if thisNodeIdx == -1 {
		panic("Could not find a node with this port in node list")
	}
	fmt.Println("initializing hash ring and membership...")
	cluster.InitMembership(nodes, nodes[thisNodeIdx], replicas)
	fmt.Println("initializing internal client...")
	client.Init(true)
	fmt.Println("starting health checks...")
	cluster.StartHealthChecks(nodes)
	fmt.Println("starting benchmarks...")
	benchmarks.StartBenchmarks()
	fmt.Println("Pulling data from peers...")
	cluster.PullFromPeers(nodes, cluster.ThisNode)

	addr := fmt.Sprintf("localhost:%d", port)
	srv := &http.Server{
		Addr:    addr,
		Handler: api.GetHandler(),
	}
	fmt.Println("serving...")
	srv.ListenAndServe()

}
