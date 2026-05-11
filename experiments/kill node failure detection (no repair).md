### Test Objective
* Test failure detection mechanism and behavior when a node dies and is brought back online later
* Gather read/write statistics during 3 time intervals: 
1. before the kill
2. after the kill but before failure is detected
3. after failure has been detected
* Gather read/write statistics along two dimensions
1. Read data that was written before the kill operation
2. Read/write new random data

### Test Setup
* Start-up healthy cluster of 3 nodes with replication factor of 3 and best_effort write mode
* Write 1000 random key/value pairs to the healthy cluster, storing the generated pairs in map M1, and tracking how many writes failed
* kill a node and do the following before failure is detected:
* read the keys in M1 from the degraded cluster tracking whether the values match
* write 1000 more random key/value pairs to the degraded cluster, storing the generated pairs in map M2, and tracking how many writes failed
* Wait until After the failure is detected by the cluster (around 12 seconds)
* read the keys in M1 from the cluster, tracking whether the values match
* read the keys in M2 from the cluster, tracking whether the values match
* write 1000 more random key/value pairs to the cluster, storing the generated pairs in map M3, and tracking how many writes failed
* restart the dead node and wait until the cluster has seen it's alive (around 4 seconds)
* read the keys in M1 from the recovered cluster, tracking whether the values match
* read the keys in M2 from the recovered cluster, tracking whether the values match
* read the keys in M3 from the recovered cluster, tracking whether the values match

## Results

### === HEALTHY CLUSTER ===
- 0.00% of puts errored
- 0.00% of gets from M1 errored
- 0.00% of gets from M1 returned incorrect value
- Interpretation: cluster is fully healthy, so no request failed

### === DEGRADED CLUSTER (pre-failure detection) ===
- 9.20% of gets from M1 errored
- 9.20% of gets from M1 returned incorrect value
- 7.80% of puts errored
- Interpretation: 1/3 nodes is dead but gets and puts are still being routed to the dead node. 

### === DEGRADED CLUSTER (POST-failure detected) ===
- 0.00% of gets from M1 errored
- 0.00% of gets from M1 returned incorrect value
- 0.00% of gets from M2 errored
- 0.00% of gets from M2 returned incorrect value
- 0.00% of puts errored
- Interpretation: The healthy nodes have detected that one node is dead, and have stopped routing traffic to it and have recalculated the hash ring. 
Data replication has ensured the data in M1 and M2 was not lost (even though 7.8% of M2 puts errored in the degraded cluster, that data was still written to the healthy replicas. This is because in best-effort mode, the coordinator may return an error even if writes succeeded on a subset of replicas. In best-effort mode, the coordinator requires that the primary succeeds in order to return success)

### === RECOVERED CLUSTER (Dead node back online, no repair) ===
- 0.00% of gets from M1 errored
- 9.20% of gets from M1 returned incorrect value
- 0.00% of gets from M2 errored
- 7.80% of gets from M2 returned incorrect value
- 0.00% of gets from M3 errored
- 8.60% of gets from M3 returned incorrect value
- Interpretation: The healthy nodes have seen the dead node is alive again and have recomputed the hash ring. However, they have not moved data to the newly alive 
node, so for the data that it is primary for, reads will return incorrect (nonexistent) values. This will be addressed in the next version of this experiment
when we enable data rebalancing. 
