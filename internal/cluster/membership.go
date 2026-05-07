package cluster

type Node struct {
	Name string
	Addr string
}

var ThisNode Node
var Replicas int

func InitMembership(nodes []Node, thisNode Node, replicas int) {
	ThisNode = thisNode
	Replicas = replicas
	rebuildHashRingWithNodes(nodes)
}

func rebuildHashRingWithNodes(nodes []Node) {
	Ring.rebuildHashRing(nodes)
	Ring.PrintHashRing()
}
