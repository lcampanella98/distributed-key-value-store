## Test objective
* Assess repair behavior and limitations when a node dies and later comes back online. When a dead node comes back up, there are a few options for how to repair the data it should contain:
1. pull the repair snapshots from peers immediately (i.e. before the peer nodes recognize the dead node is healthy)
2. pull the repair snapshots from peers after it has rejoined the cluster

* The hypothesis is it's better to pull the repair snapshots after it has rejoined the cluster. Reason being, if we repair immediately on node recovery, writes that occur between that time and the time the other nodes see the recovered node as alive again and begin routing traffic to it, will never make it onto the recovered node, causing stale reads. But Delaying repair until after we're confident the node has rejoined the cluster accounts for those writes in the repair. 
* Note how reads currently behave in the system: a get request attempts to read only from the primary. So if the network request to primary failed, or the requested data on the primary is stale, the get request will fail or contain stale data. This pessimistic consistency model was chosen right now for its simplicity. 


## Test Setup
* We constantly put random pairs to the cluster every millisecond (1000 puts/second)
* Each time we put a random pair to the cluster, 5 seconds later we perform a get of that key and compare the retrieved value to the expected value. Note that different delays would produce different stale read percentages, so results are dependent on this delay. 
* For the first 5 seconds, the cluster is healthy with 3 nodes and a replication factor of 3
* Then we kill a node, the node is dead for the next 17 seconds
* Then we revive the node, the node is alive for the next 15 seconds, at which point we end the test. 
* All this time the system is performing the get/put logic described above, and we track the failed gets and puts within each time window
* We run this test three times. Once with no repair (baseline), one with immediate repair, and one with delayed repair. 


## No Repair Results (Baseline)
- === HEALTHY CLUSTER ===
- 0.00% of puts errored (0)
- 0.00% of gets errored (0)
- 0.00% of gets returned incorrect value (0)
- === 1 NODE DEAD CLUSTER ===
- 1.47% of puts errored (250)
- 1.90% of gets errored (317)
- 1.90% of gets returned incorrect value (317)
- === NODE RECOVERED CLUSTER ===
- 0.00% of puts errored (0)
- 0.00% of gets errored (0)
- 3.35% of gets returned incorrect value (454)

## Immediate Repair Results
- === HEALTHY CLUSTER ===
- 0.00% of puts errored (0)
- 0.00% of gets errored (0)
- 0.00% of gets returned incorrect value (0)
- === 1 NODE DEAD CLUSTER ===
- 1.92% of puts errored (230)
- 2.09% of gets errored (245)
- 2.09% of gets returned incorrect value (245)
- === NODE RECOVERED CLUSTER ===
- 0.00% of puts errored (0)
- 0.00% of gets errored (0)
- 1.08% of gets returned incorrect value (146)

## Delayed Repair Results
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

## Interpretation
* We are particularly interested in the differences after the dead node came back online, i.e. "NODE RECOVERED CLUSTER" sections. 
* First observation, The "No Repair" test had far more stale reads after node recovery than the tests with repair enabled. This makes sense as no data was moved onto the recovered node, but reads (& writes) were now being routed to it. 
* Next important observation: The stale reads after node recovery were not significantly different between Immediate repair vs. Delayed repair. This does not line-up with our hypothesis, which recall was that Delayed Repair should perform **better** than Immediate Repair. 
* Why were there still a significant number of stale reads in Delayed Repair? Considering that Delayed Repair should account for writes that occurred between the dead node recovering and when it rejoins the cluster? 
* What we failed to consider is that in Delayed Repair, the node always rejoins the cluster slightly **before** repair snapshots are pulled, meaning that there is a time interval during which reads are routed to the recovered node, but repair has **not yet happened**, resulting in stale reads. 
* How to address these stale reads? 
* Unfortunately there is no great way to ensure repair runs at exactly the same time the node rejoins the cluster, considering health checks are periodic. Even if there were a way, we would still encounter some stale reads (though fewer than observed in this experiment)
* The best way to address this issue is probably not chasing exact timing of the repair, but instead implementing read from replicas and/or read repair. 