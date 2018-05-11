# Clustering

The Event Gateway is a horizontally scalable system. It can be scaled by adding instances to the cluster. A cluster is
a group of instances sharing the same database. A cluster can be created in one cloud region, across multiple regions,
across multiple cloud provider or even in both cloud and on-premise data centers.

The Event Gateway instances use a strongly consistent, subscribable DB (initially [etcd](https://coreos.com/etcd),
with support for Consul, and Zookeeper planned) to store and broadcast configuration. The instances locally
cache configuration used to drive low-latency event routing. The instance local cache is built asynchronously based on
events from backing DB.

The Event Gateway is a stateless service and there is no direct communication between different instances. All
configuration data is shared using backing DB. If the instance from region 1 needs to call a function from region 2 the
invocation is not routed through the instance in region 2. The instance from region 1 invokes the function from region 2
directly.

```
┌─────────────────────────────────────────────Event Gateway Cluster──────────────────────────────────────────────┐
│                                                                                                                │
│                                                                                                                │
│                                            Cloud Region 1───────┐                                              │
│                                            │                    │                                              │
│                                            │   ┌─────────────┐  │                                              │
│                                            │   │             │  │                                              │
│                   ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│─ ▶│etcd cluster │◀ ┼ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─                    │
│                                            │   │             │  │                          │                   │
│                   │                        │   └─────────────┘  │                                              │
│                                            │          ▲         │                          │                   │
│                   │                        │                    │                                              │
│        Cloud Region 2───────┐              │          │         │               Cloud Regio│ 3───────┐         │
│        │          │         │              │                    │               │                    │         │
│        │          ▼         │              │          ▼         │               │          ▼         │         │
│        │  ┌───────────────┐ │              │  ┌──────────────┐  │               │  ┌──────────────┐  │         │
│        │  │               │ │              │  │              │  │               │  │              │  │         │
│        │  │ Event Gateway │ │              │  │Event Gateway │  │               │  │Event Gateway │  │         │
│        │  │   instance    │◀┼──────────┐   │  │   instance   │◀─┼──────────┐    │  │   instance   │  │         │
│        │  │               │ │          │   │  │              │  │          │    │  │              │  │         │
│        │  └───────────────┘ │          │   │  └──────────────┘  │          │    │  └──────────────┘  │         │
│        │          ▲         │          │   │          ▲         │          │    │          ▲         │         │
│        │          │         │          │   │          │         │          │    │          │         │         │
│        │          │         │          │   │          │         │          │    │          │         │         │
│        │          ▼         │          │   │          ▼         │          │    │          ▼         │         │
│        │        ┌───┐       │          │   │        ┌───┐       │          │    │        ┌───┐       │         │
│        │        │ λ ├┐      │          └───┼───────▶│ λ ├┐      │          └────┼───────▶│ λ ├┐      │         │
│        │        └┬──┘│      │              │        └┬──┘│      │               │        └┬──┘│      │         │
│        │         └───┘      │              │         └───┘      │               │         └───┘      │         │
│        └────────────────────┘              └────────────────────┘               └────────────────────┘         │
│                                                                                                                │
│                                                                                                                │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```
