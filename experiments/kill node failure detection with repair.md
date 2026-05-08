## Test Objective
* Test data repair behavior when a node dies and is brought back online some time later
* This test is identical to the "kill node failure detection" test, but with repair enabled
* Repair means when the dead node is restarted and rejoins the cluster, it pulls the required data from its peers

## Test setup
* Identical to test setup of "kill node failure detection"

## Results

### === HEALTHY CLUSTER ===
- 0.00% of puts errored
- 0.00% of gets from M1 errored
- 0.00% of gets from M1 returned incorrect value
- Interpretation: cluster is fully healthy, so no request failed


### === DEGRADED CLUSTER (PRE-failure detection) ===
- 8.70% of gets from M1 errored
- 8.70% of gets from M1 returned incorrect value
- 9.20% of puts errored
- Interpretation: 1/3 nodes is dead but gets and puts are still being routed to the dead node. 

### === DEGRADED CLUSTER (POST-failure detected) ===
- 0.00% of gets from M1 errored
- 0.00% of gets from M1 returned incorrect value
- 0.00% of gets from M2 errored
- 0.00% of gets from M2 returned incorrect value
- 0.00% of puts errored
- Interpretation: The healthy nodes have detected that one node is dead, and have stopped routing traffic to it and have recalculated the hash ring. 
Data replication has ensured the data in M1 and M2 was not lost (even though 7.8% of M2 puts errored in the degraded cluster, that data was still written to the healthy replicas. This is because in best-effort mode, the coordinator may return an error even if writes succeeded on a subset of replicas. In best-effort mode, the coordinator requires that the primary succeeds in order to return success)

### === RECOVERED CLUSTER (Dead node back online) ===
- 0.00% of gets from M1 errored
- 0.00% of gets from M1 returned incorrect value
- 0.00% of gets from M2 errored
- 0.00% of gets from M2 returned incorrect value
- 0.00% of gets from M3 errored
- 0.00% of gets from M3 returned incorrect value
- Interpretation: When the dead node was brought back online, it pulled the required data from its peers. The peers calculated the data that the restarted node should have, i.e. the data for which the restarted node is part of the key's replica set, and returned that subset of data to the restarted node. The healthy nodes then see the dead node is alive again, recompute the hash ring, and begin routing traffic to it again. In contrast to the "kill node failure detection (no repair)" test, all get operations to the recovered cluster return the correct value. 