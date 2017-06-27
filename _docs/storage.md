# Implications of Storage Semantics

For configuration, this system has this desired behavior:

1. something wants to configure something
1. everything that needs to know about that configuration reacts to it

We benefit from strongly consistent atomic writes because we want all configuration to
reach everything that needs to be configured. Weakly consistent systems
may not observe all updates, and we will end up with broken or half-working
features. We also benefit from efficient notification primitives for
keeping the overall system performant, such as watches. Without watches,
we have to continuously scan ranges of keys in the database until we
find new data, which is enormously expensive, slow, and hard to scale.

Applied to security, where we want to revoke a user's privileges completely,
if we are using a database without the ability to perform atomic writes,
even if it is strongly consistent,
then one component may try updating a user's account by first reading
their data, modifying it locally, then trying to update it in the database.
If we don't have atomic updates, the ordering could end up being this:

```
node A reads user Bob's data
node A locally updates Bob's data to add admin privileges
node B deletes Bob's data to remove all privileges
node B receives "write successful" from database
node A writes Bob's new data to the system with admin privileges
Node B then tells the system that was trying to delete Bob "Bob sucessfully deleted"
Bob steals all of the company's secrets
```

This is not desirable. However, some strongly consistent databases let us
perform compare-and-swap operations that let us atomically update Bob, without
losing any intermediate updates.

```
node A reads user Bob's data
node A locally updates Bob's data to add admin privileges
node B deletes Bob's data to remove all privileges
node B receives "write successful" from database
node A tries to do "update Bob unless changed since read" which fails
Node B then tells the system that was trying to delete Bob "Bob sucessfully deleted"
Bob is locked out and steals nothing
```

Atomic updates generally significantly reduce the cognitive burden for
building a system which must store and react to state changes in different
components. Watches are another significant help for this, as they
notify interested systems in relevant changes.

Watches are how we apply the event-driven model to our database.
When interesting changes happen, interested parties react to them.
This decouples the emitter from the reactor, greatly simplifying
interactions in the system. Without watches, we need to
have some way of detecting changes in a database. If we can scan through
all keys, we can do an O(N) traversal of the entire database, which
does not scale very high, but could work alright. If we don't have
the ability to scan through all data, we may actually have no way of
learning what changes are unless there is a top-level key that is set
that holds everything. This does not scale beyond a couple kilobytes.

# Storage Options

## stateless gateway services backed by zk/etcd/consul cluster, abstracted by docker/libkv

pros

* flexible support for the three most popular configuration databases
* very clear operational characteristics, does not confuse anyone about what's happening
* the gateway is fully stateless, easy to autoscale, clear semantics for operators
* lowest amount of work for us
* write once, run anywhere in the same way
* allows users to easily take advantage of existing database skills, tools, backup tools, monitoring, etc...

cons

* requires users to run their own cluster (they are already running things though, so this isn't a high marginal cost)

## embedded etcd in gateway, cluster of 3 or 5 active as "leaders", rest of gateways are stateless

pros

* single binary

cons

* unclear operational characteristics, when the cluster gets wedged it may be extremely hard to debug
* harder than running your own cluster, because you can't reuse existing database skills
* very hard on operators when things go wrong
* harder on operators to get things safely set up
* still effectively have a separate cluster if you want to autoscale without accidentally losing leaders
* creating a reliable "autopilot" etcd deployment system took tyler 4 months in the past

## embedded etcd in gateway, single node configured as "leader", rest of gateways are stateless

pros

* easy to set up
* can be used in combination with a separate etcd cluster for the best of both worlds
* does not require setting up a cluster to try out, demo, or run in small deployments
* clear operational characteristics, everyone knows it's not reliable

cons

* single point of failure on the single leader node

## dynamo + spanner + cosmosdb + on prem other databases

pros

* single binary
* easy deployment for users on cloud providers
* one fewer piece to keep running

cons

* dramatically increases the complexity for building the system for us
* we need to target multiple consistency models (hard+++)
* need to target systems without watch semantics (hard+++)
* need to target systems without atomic write semantics (hard+++)
* we need to spend much more effort on testing
* we need to spend much more effort on fixing bugs
* we need to spend much more effort on writing monitoring code
* we need to master all of these databases in order to program against them

## eventually consistent gossip-based config sharing with CRDTs

pros

* single binary
* no external dependencies, even on databases

cons

* if all instances go away, the configuration is gone
* we're basically building our own distributed database (hard+++++)
* need to create backup tooling
* unclear operational characteristics for people
* huge extra effort needs to go into correctness testing

## just config files, reloaded on file change

pros

* single binary
* flexible
* can be used with any of the other approaches to decouple functionality

cons

* users don't get an API for the gateway

# Recommendation

Keep the system stateless to make it easy to operate. Store state
in a database that supports atomic updates and watches, such as
etcd, zookeeper, or consul. Use the docker/libkv library to
support all 3 backing databases. This lets us treat every
deployment environment the same way, regardless of whether
it's in a cloud provider or on-prem. This combination of
operational clarity and database semantics will significantly
lower the amount of effort we need to put into engineering over time.

Purely for demo and trial purposes, allow a gateway to be started
with an `--embedded-master` flag which will start an embedded etcd instance
that other gateways can use as a shared backing store. This allows people to
try out the system without standing up an etcd cluster first.

The main downside, having to run a separate cluster in production, is probably
not a big deal for people who are interested in running this themselves.

Assumptions relating to this cost, and it not seeming very high:

1. most users will fall into these camps:
  * small/medium orgs interested in trying out locally but using the SaaS offering
  * orgs with more significant engineering resources who already run zk/etcd/consul
  * orgs with more significant engineering resources who are not averse to running zk/etcd/consul
  * orgs who are comfortable paying compose.com $30/mo for hosted etcd
1. users who are not interested in paying for SaaS but still want to run the gateway
   themselves WITHOUT running a database are unlikely to be very upset when they occasionally
   need to reconfigure the gateway when the single master goes away.