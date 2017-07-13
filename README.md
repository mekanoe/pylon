# Pylon

Distributing read-heavy Redis systems for global-scale master-master replication.

## Architecture

Pylon works in 3 primary layers and a secondary clustering layer, all of which are transparent and can be ran in a single process, or completely separate reader-writer processes. Pylon sites on top of many individual Redis clusters that aren't connected, and keeps them in sync. Any Redis operation done is only to a Pylon's local Redis.

- **Gateway Layer** - This layer speaks Redis to the application, and decides how to route a request. This is a code-level read/write splitting proxy, that'll just execute a callback if it's a read or a write. In the full scope, a read is directed directly into the Redis Layer and it's return value is returned to the requestor ASAP; while writes are confirmed after being pushed to the Stream Layer (even if Redis doesn't ultimately accept it, **this is an important implementation detail**) 

- **Stream Layer** - This layer talks to Kafka. In all situations, this is where writes go and these operations are appended to Kafka's commit log. In situations where the current Pylon is allowed to write to Redis (a rack leader,) the Kafka streamed operations are sent to the Redis Layer for handling. The significance of Kafka is that global consistency is handled by Kafka itself, and since we can arbitrarily tell Kafka what position we're in for new commits, catching up a new cluster or recovering a failed one is a first-class operation.

- **Redis Layer** - This layer talks to Redis. It can use any Redis-speaking system underneath it (e.g. ElastiCache, twemproxy, ssdb, Redis Cluster, CurioDB, ledis) as well as get cluster configuration from Sentinels. It could, in theory, be used for a heterogenous system where if a standalone redis instance needs to be sharded into a redis-cluster or twemproxy, a new Pylon rack could be created, caught up, and the old Pylon rack could be removed on the fly. It could even do something silly, like talking to another Pylon (although doing this to an interconnected Pylon will possibly explode Kafka and fill your Redis AOFs into something astronomical.)

- **Clustering Layer** - This layer talks to the local rack as well as the world to figure out two things: who will write, and where do we read. Pylon is rack-aware, so in an HA situtation, only one Pylon will ever be writing to it's local Redis. The world-awareness is to make sure reads **always** happen if there is a Redis online. Periodically, the Pylons will congregate and figure out the fastest path between each other in case of failover, and if the local Redis is away, it'll quickly pipe the read to the next fastest Pylon, and your read will happen. In case of writes, the local Redis can be away, but since incoming write operations are going into Kafka, a Pylon that can't talk to it's local Redis can still write for the world, and the Redis will catch up when it returns. This layer also will manage voting in the case of contentious keys.

Any operation that both reads and modifies data will not work. Pubsub will not work, however, you do have a huge Kafka commit log you could process for the same data with Samza or something similar; or for non-key-cache pubsub, you have Kafka in general. Why not use it?!

In the case of write contention, the last writer always wins, unless a key is marked in configuration as "contentious." If a key is contentious and needs quorum, the world will vote upon who wins, or if a merge shall occur. If no keys are marked as contentious, the Clustering layer's Raft will rarely be used outside of leader election. 

*Notes: A "rack" is any individual cluster, where a pylon and optionally it's underlying redis cluster sit, and the "world" is the global cluster of many racks. Racks have individual leaders and the world also has a leader.

## Developing

For regular development, use `docker-compose up`. If you need a proper cluster, use `docker-compose up -f docker/dev-cluster.yml`.

## Redis Command Notes

Most Redis commands work fine, however, it's imperative that you understand a few quirks with commands, and you must work around these:

### Blacklisted Commands

- All Pubsub commands aren't allowed. This may be implemented leveraging Kafka instead in the future. Alternatively, if you are listening for changes on a key, use the underlying Redis directly if possible, as the changes will still propagate through the world.

- All blocking operations aren't allowed, such as `BLPOP`. Pylon doesn't wait for anything. 

- Transations aren't allowed. (I promise, Pylon can handle a crazy number of writes at a time and they'll end up in Redis at some point.)

- Scripting isn't allowed.

- Cluster commands should be directed at the underlying Redis. Pylon doesn't know what to do about these.

- `AUTH` isn't allowed, although this may be implemented in the future.

- `SELECT` is silently ignored. Pylon can be configured to select a backing database, but Pylon is always on index 0, and has no notion of selecting.

- `GETSET` is blacklisted, use a `GET` then `SET` instead. This may be transparently done in the future.


### Dangerous & Noteworthy Commands

- `SORT` must not also include a `STORE` directive. `SORT` is handled as a transparent read operation so this will **break** consistency.

- `LPOP` nor any similar `*POP*` command will not return a value, as it's handled as a write operation. If you need the last value, do `LINDEX listkey -1` to read then `LPOP listkey`, or similar flow. This may be transparently done in the future.

### Pylon Internal Commands

- `PYLON INFO` - Returns info about the Pylon.