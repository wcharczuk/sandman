```
                             __
    _____ ____ _ ____   ____/ /____ ___   ____ _ ____ 
   / ___// __ `// __ \ / __  // __ `__ \ / __ `// __ \
  (__  )/ /_/ // / / // /_/ // / / / / // /_/ // / / /
 /____/ \__,_//_/ /_/ \__,_//_/ /_/ /_/ \__,_//_/ /_/ 
```

`sandman` is a an experimental "timer" orchestrator.

Think of it as an answer to the question; how would we handle needing to send ~1 million webhooks at a specific point in the future within the same minute?

# Goals

The design goals with `sandman` are as follows:

1. Support sending ~1 million timer "hooks" per minute at arbitrary points in the future.
2. Support some fixed nominal retry count for each hook.
3. Support basic "priorities" with timers such that we can customize the order that the timers are processed.

# General Approach

`sandman` leans on two basic concepts to achieve scale
- [Hashed and hierarchical timing wheels](https://dl.acm.org/doi/10.1145/41457.37504) to structure the timers
- [CockroachDB](https://www.cockroachlabs.com/) to scale the database layer horizontally

Typical implementations of calendarized event tables use indexes and timestamps to organize the data in a way that can be filtered quickly. This can work well in practice for even large counts of timers, but as the table(s) that back these timers grows it gets slower and slower to insert new timers because table indexes need to be updated for each timer inserted.

`sandman` echews indexed timestamps, and instead uses "counter" fields that are updated every minute by a scheduler process which represent the minutes until the timer is due. 

For example, if you have a timer that is due in 2 years, you would put (2*365*24*60 == 1,051,200 minutes) as the counter value, and it would be decremented each cycle for 2 years. 

In practice, a scheduler pool picks a leader, and then that leader every minute runs a very simple update command on the entire database to decrement counters. This sounds expensive, but remember that we can scale CockroachDB horizontally to make this update command relatively performant. We do this by splitting the timer table up across N nodes, where each node has some subset of the total timers. When we run the update, each node runs the update on its subset of data, and if we need to handle more timers, we simply add more nodes. 

Similarly, to detect if timers are "due" we simply do a select on each node for timers with counter==0, which in practice scans the table for rows that match the predicate, again leaning on the idea that we can add more nodes to make this efficient if necessary.

To deliver the timer "hooks" we have a worker pool that can scale as needed, where each worker pool every 10 seconds scans for new work, and attempts to deliver the hooks in parallel.

# Getting started

1. Make sure prerequisites are installed
  1a. See installing [CockroachDB locally]()
  1b. Make sure that the [protobuf compiler]() is installed locally
2. Make sure protobuf plugins are installed:
```bash
> make init
```
3. Create the underlying database and run migrations
```bash
> make db
```
4. Install the management cli (`sandctl`)
```bash
> go install sandman/sandctl
```
5. Start the local cluster
```bash
> make run
```

While the cluster is running, in another terminal window you can perform additional steps.

For instance you can now create a timer, and for convenience there is a "generator" in `sandctl` which can generate the yaml format that `sandctl timer create` expects:

```bash
> sandctl timer generate --name=$(uuidgen) --due-in=10m --hook-url=http://localhost:8081/foo/bar --priority 1000 --hook-method=GET --label=env=prod --label=cluster=northwest | sandctl timer create -f -
```

You just created your first timer! In about 10 minutes `sandman` will try and deliver it.
