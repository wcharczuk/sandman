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
- [Shuffle Sharding](https://aws.amazon.com/builders-library/workload-isolation-using-shuffle-sharding/) to distribute timers fairly
- [CockroachDB](https://www.cockroachlabs.com/) to scale the database layer horizontally

When a timer is created, it gets inserted into one big table, the `timers` table, with a `due_utc` value. The `timers` table has no indexes outside a unique key for the user supplied "name" field.

From here, workers poll the table every 10 seconds, pulling 255 "due" timers per poll. These timers are sent via. their "hook" details with 5 second timeouts.

Workers can be added as needed to scale for timer due spikes, as they cooperatively mark subsets of the `timers` table for assignment.

Similarly, as the database becomes loaded, additional replicas can be added to distribute shares of the `timers` table to new nodes such that the table scans for the polls do not exceed some nominal thresholds.

The idea here is by keeping the table simple, and keeping inserts fast and leaning on horizontal scale for polling and delivery, we can scale the system as needed to handle load.

Additionally, timers have `shard_key` fields to provide a mechanism to distribute timers "fairly" based on shard assignment. The way this works in practice is when workers poll for timers, they order by the user-supplied priority on the timer, and add a "boost" priority based on the shard key assignment within 3600 buckets (i.e. seconds within the hour). This has the effect of randomizing the polling order and cutoff for which timers are fetched for each due period, keeping "hot" shards from hogging the delivery capacity.

# Getting started

1. Make sure prerequisites are installed:

  1a. See installing [CockroachDB locally](https://www.cockroachlabs.com/docs/v24.2/install-cockroachdb-mac)

  1b. Make sure that the [protobuf compiler](https://grpc.io/docs/protoc-installation/) is installed locally

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
