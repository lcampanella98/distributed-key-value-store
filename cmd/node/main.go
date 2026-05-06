package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/lcampanella98/distributed-key-value-store/internal/api"
	"github.com/lcampanella98/distributed-key-value-store/internal/cluster"
)

func main() {
	nodesPtr := flag.String("nodes", "", "pass a comma separated list of node hostnames")
	portPtr := flag.Int("port", 0, "pass a port for this node to run on")
	flag.Parse()
	nodeNames := strings.Split(*nodesPtr, ",")
	port := *portPtr
	nodes := []cluster.Node{}
	var thisNode cluster.Node
	for _, name := range nodeNames {
		node := cluster.Node{Name: name, Addr: "http://" + name}
		nodes = append(nodes, node)
		if strings.Contains(name, strconv.Itoa(port)) {
			fmt.Printf("Found this node as %s\n", node.Name)
			thisNode = node
		}
	}

	cluster.InitHashRing(nodes, thisNode)
	cluster.PrintHashRing()

	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:    addr,
		Handler: api.GetHandler(),
	}
	fmt.Println("serving...")

	srv.ListenAndServe()

}
