### Test Objective
* Test strict write mode vs best_effort write mode
* strict write mode chooses consistency over availability, where successful puts require successful responses from all replicas
* best_effort write mode chooses availability over consistency, where successful puts only require that the primary succeeded. In best-effort mode, the coordinator may return an error even if writes succeeded on a subset of replicas

### Test setup for each mode
* run 3 nodes and replication factor of 3 (all nodes on the hash ring store a copy of the data)
* choose a coordinator node
* kill a node that is not the coordinator node
* Run using testing/error_statistics.go which does the following
* for 1000 random key/value pairs:
* put the key/value pair
* get the key
* check whether the returned value == expected value

## Results

### strict write mode (successful puts require successful responses from all replicas, i.e. consistency over availability)
- 100.00% of puts errored
- 8.90% of gets errored
- 8.90% of gets returned incorrect value
- Interpretation: all puts errored because a successful put requires successful responses from all 3 nodes, but only 2 nodes were up. 
only a fraction of gets errored because the put operations still went through on the 2 healthy nodes. the dead node was primary for only that fraction of the keys (ideally that fraction should be closer to 1/3 and can be improved with virtual nodes on the hash ring)

### best_effort write mode (successful puts only require that the primary succeeded)
- 7.60% of puts errored
- 7.60% of gets errored
- 7.60% of gets returned incorrect value
- Interpretation: only a fraction of puts errored this time because a successful put here only requires that the primary succeeded. The dead node was primary for only that fraction of the pairs. 