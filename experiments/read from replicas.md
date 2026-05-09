## Test Objective
* Compare system behavior during node failures under two scenarios:
1. Read from primary only
2. Read from primary + replicas
* To elaborate on the 2nd one, we will first attempt to read from primary, and if that fails, we attempt to read from the next replica, and so on. We only fail the request if all replicas fail the read. 
* The hypothesis is that availability of reads will be higher when we read from replicas

## Test Setup
* (Same as Test Setup in "Repair snapshot inconsistency" experiment) Reproduced below:
* We constantly put random pairs to the cluster every millisecond (1000 puts/second)
* Each time we put a random pair to the cluster, 5 seconds later we perform a get of that key and compare the retrieved value to the expected value. Note that different delays would produce different stale read percentages, so results are dependent on this delay. 
* For the first 5 seconds, the cluster is healthy with 3 nodes and a replication factor of 3
* Then we kill a node, the node is dead for the next 17 seconds
* Then we revive the node, the node is alive for the next 15 seconds, at which point we end the test. 
* All this time the system is performing the get/put logic described above, and we track the failed gets and puts within each time window


## Test Evaluation
* Run the test two times, one with read from primary only, and one with read from replicas

## Results Read from primary only
- === HEALTHY CLUSTER ===
- 0.00% of puts errored (0)
- 0.00% of gets errored (0)
- 0.00% of gets returned incorrect value (0)
- === 1 NODE DEAD CLUSTER ===
- 1.42% of puts errored (242)
- 2.27% of gets errored (378)
- 2.27% of gets returned incorrect value (378)
- === NODE RECOVERED CLUSTER ===
- 0.00% of puts errored (0)
- 0.00% of gets errored (0)
- 1.12% of gets returned incorrect value (152)

## Results Read from Replicas
- === HEALTHY CLUSTER ===
- 0.00% of puts errored (0)
- 0.00% of gets errored (0)
- 0.00% of gets returned incorrect value (0)
- === 1 NODE DEAD CLUSTER ===
- 1.17% of puts errored (199)
- 0.00% of gets errored (0)
- 0.00% of gets returned incorrect value (0)
- === NODE RECOVERED CLUSTER ===
- 0.00% of puts errored (0)
- 0.00% of gets errored (0)
- 0.95% of gets returned incorrect value (128)
- Node recovered at: 2026-05-09 11:12:05.726287 -0400 EDT m=+34.122146301
- First stale read: 2026-05-09 11:12:08.6946982 -0400 EDT m=+37.090557501 and stale read window was 2.073896s


## Metrics Read from Replicas from coordinator
- === Metrics ===
- puts_total=32000
- puts_2xx=31801
- puts_4xx=0
- puts_5xx=199
- puts_failed=199
- avg_put_latency_ms=0.04
- gets_total=32000
- gets_2xx=32000
- gets_4xx=0
- gets_5xx=0
- gets_failed=0
- avg_get_latency_ms=0.01
- replication_requests_total=47341
- replication_failed=2443
- avg_replication_latency_ms=0.03
- get_from_replica_requests_total=26754
- get_from_replica_failed=244
- avg_get_from_replica_latency_ms=0.01
- repair_requests_total=3
- repair_failed=0
- avg_repair_latency_ms=5.67
- alive_nodes=2
- ring_rebuilds=4
- repair_keys_transferred=25567


## Interpretation
* We see that in the time window when the node was dead read from replicas had 100% availability and no incorrect reads, confirming our hypothesis. Note that this was only under single-node failure conditions
* Although this is a strong result for the our availability, it moves the system into more consistency risks. If the primary is down and the replica contains stale data we would surface that stale data to the client. The impact of this depends on the use case and how acceptable stale reads are. 
* We also didn't see a significant difference in stale reads during node recovery. The stale reads were the result of repair snapshots happening slightly after the recovered node rejoins the cluster. This is confirmed by the timestamps of first and last stale read compared with the timestamp of node recovery. 
* How to address this stale read window on node recovery? 
* One possible and simple solution is slightly modifying the read from replicas logic to continue searching replicas as long as they fail OR report key doesn't exist. This would solve the issue because for each stale read in that stale read window the recovered node simply doesn't have the data yet, and flags the nonexistant key in its response. This modified logic will not cause a problem because there is no delete key API, so we know that a nonexistant key either means the key was never put, or the read was stale. BUT if we added a delete API later, or a TTL feature, this optimization becomes unsafe immediately. 
* Below are the results with this "skip nonexistant keys" logic below. As expected, we see 0 stale reads during node recovery. 

## Results Read from Replicas (Skip nonexistant keys)
- === HEALTHY CLUSTER ===
- 0.00% of puts errored (0)
- 0.00% of gets errored (0)
- 0.00% of gets returned incorrect value (0)
- === 1 NODE DEAD CLUSTER ===
- 1.22% of puts errored (207)
- 0.00% of gets errored (0)
- 0.00% of gets returned incorrect value (0)
- === NODE RECOVERED CLUSTER ===
- 0.00% of puts errored (0)
- 0.00% of gets errored (0)
- 0.00% of gets returned incorrect value (0)
- Node recovered at: 2026-05-09 12:06:25.8760874 -0400 EDT m=+34.980778401
