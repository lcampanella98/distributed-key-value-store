package cluster

type Node struct {
	Name string
	Addr string
}

var ThisNode Node
var Replicas int
var AllNodes []Node

func InitMembership(nodes []Node, thisNode Node, replicas int) {
	AllNodes = nodes
	ThisNode = thisNode
	Replicas = replicas
	rebuildHashRingWithNodes(nodes)
}

func rebuildHashRingWithNodes(nodes []Node) {
	Ring.rebuildHashRing(nodes)
	Ring.PrintHashRing()
}
