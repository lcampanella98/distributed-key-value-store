package mymetrics

import (
	"fmt"
	"sync"
	"time"
)

type Metrics struct {
	mu sync.Mutex

	putsTotal       int64
	puts2xx         int64
	puts4xx         int64
	puts5xx         int64
	putLatencyTotal time.Duration

	getsTotal       int64
	gets2xx         int64
	gets4xx         int64
	gets5xx         int64
	getLatencyTotal time.Duration

	replicationRequestsTotal int64
	replicationFailed        int64
	replicationLatencyTotal  time.Duration

	getFromReplicaTotal        int64
	getFromReplicaFailed       int64
	getFromReplicaLatencyTotal time.Duration

	repairRequestsTotal int64
	repairFailed        int64
	repairLatencyTotal  time.Duration

	aliveNodes            int64
	ringRebuilds          int64
	repairKeysTransferred int64
}

var M Metrics

func (m *Metrics) StartMetricDumper(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			m.Dump()
		}
	}()
}

func (m *Metrics) Dump() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var avgPutLatencyMs float64
	if m.putsTotal > 0 {
		avgPutLatencyMs = float64(m.putLatencyTotal.Milliseconds()) / float64(m.putsTotal)
	}

	var avgGetLatencyMs float64
	if m.getsTotal > 0 {
		avgGetLatencyMs = float64(m.getLatencyTotal.Milliseconds()) / float64(m.getsTotal)
	}

	var avgReplicationLatencyMs float64
	if m.replicationRequestsTotal > 0 {
		avgReplicationLatencyMs = float64(m.replicationLatencyTotal.Milliseconds()) / float64(m.replicationRequestsTotal)
	}

	var avgRepairLatencyMs float64
	if m.repairRequestsTotal > 0 {
		avgRepairLatencyMs = float64(m.repairLatencyTotal.Milliseconds()) / float64(m.repairRequestsTotal)
	}

	var avgGetFromReplicaLatencyMs float64
	if m.getFromReplicaTotal > 0 {
		avgGetFromReplicaLatencyMs = float64(m.getFromReplicaLatencyTotal.Milliseconds()) / float64(m.getFromReplicaTotal)

	}

	fmt.Println("=== Metrics ===")

	fmt.Printf("puts_total=%d\n", m.putsTotal)
	fmt.Printf("puts_2xx=%d\n", m.puts2xx)
	fmt.Printf("puts_4xx=%d\n", m.puts4xx)
	fmt.Printf("puts_5xx=%d\n", m.puts5xx)
	fmt.Printf("puts_failed=%d\n", m.puts4xx+m.puts5xx)
	fmt.Printf("avg_put_latency_ms=%.2f\n", avgPutLatencyMs)

	fmt.Printf("gets_total=%d\n", m.getsTotal)
	fmt.Printf("gets_2xx=%d\n", m.gets2xx)
	fmt.Printf("gets_4xx=%d\n", m.gets4xx)
	fmt.Printf("gets_5xx=%d\n", m.gets5xx)
	fmt.Printf("gets_failed=%d\n", m.gets4xx+m.gets5xx)
	fmt.Printf("avg_get_latency_ms=%.2f\n", avgGetLatencyMs)

	fmt.Printf("replication_requests_total=%d\n", m.replicationRequestsTotal)
	fmt.Printf("replication_failed=%d\n", m.replicationFailed)
	fmt.Printf("avg_replication_latency_ms=%.2f\n", avgReplicationLatencyMs)

	fmt.Printf("get_from_replica_requests_total=%d\n", m.getFromReplicaTotal)
	fmt.Printf("get_from_replica_failed=%d\n", m.getFromReplicaFailed)
	fmt.Printf("avg_get_from_replica_latency_ms=%.2f\n", avgGetFromReplicaLatencyMs)

	fmt.Printf("repair_requests_total=%d\n", m.repairRequestsTotal)
	fmt.Printf("repair_failed=%d\n", m.repairFailed)
	fmt.Printf("avg_repair_latency_ms=%.2f\n", avgRepairLatencyMs)

	fmt.Printf("alive_nodes=%d\n", m.aliveNodes)
	fmt.Printf("ring_rebuilds=%d\n", m.ringRebuilds)
	fmt.Printf("repair_keys_transferred=%d\n", m.repairKeysTransferred)
}

func (m *Metrics) IncPutsTotal() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.putsTotal++
}

func (m *Metrics) IncPuts2xx() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.puts2xx++
}

func (m *Metrics) IncPuts4xx() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.puts4xx++
}

func (m *Metrics) IncPuts5xx() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.puts5xx++
}

func (m *Metrics) ObservePutLatency(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.putLatencyTotal += d
}

func (m *Metrics) IncGetsTotal() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getsTotal++
}

func (m *Metrics) IncGets2xx() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gets2xx++
}

func (m *Metrics) IncGets4xx() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gets4xx++
}

func (m *Metrics) IncGets5xx() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gets5xx++
}

func (m *Metrics) ObserveGetLatency(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getLatencyTotal += d
}

func (m *Metrics) IncReplicationRequestsTotal() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.replicationRequestsTotal++
}

func (m *Metrics) IncReplicationFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.replicationFailed++
}

func (m *Metrics) ObserveReplicationLatency(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.replicationLatencyTotal += d
}

func (m *Metrics) IncGetFromReplicaRequestsTotal() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getFromReplicaTotal++
}

func (m *Metrics) IncGetFromReplicaFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getFromReplicaFailed++
}

func (m *Metrics) ObserveGetFromReplicaLatency(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getFromReplicaLatencyTotal += d
}

func (m *Metrics) IncRepairRequestsTotal() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.repairRequestsTotal++
}

func (m *Metrics) IncRepairFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.repairFailed++
}

func (m *Metrics) ObserveRepairLatency(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.repairLatencyTotal += d
}

func (m *Metrics) SetAliveNodes(n int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.aliveNodes = n
}

func (m *Metrics) IncRingRebuilds() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ringRebuilds++
}

func (m *Metrics) AddRepairKeysTransferred(n int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.repairKeysTransferred += n
}
