# Collecting and Viewing Event Gateway Metrics

This guide details how to collect and analyze metrics from the Event Gateway.

## Contents
1. [Introduction](#introduction)
1. [Installing Prometheus](#installing-prometheus)
1. [Collecting Metrics](#collecting-metrics)
1. [Visualizing Data](#visualizing-data)
1. [List of Metrics](#list-of-metrics)
    1. [etcd cluster](#etcd-cluster)
    1. [gateway](#gateway)
    1. [go](#go)

### Introduction

The Event Gateway exposes a number of [Prometheus](https://prometheus.io/)-based metrics + counters via the configuration 
API to help monitor the health of the gateway. The table below outlines the specific metrics available to query from the endpoint:

### Installing Prometheus

### Collecting Metrics

### Visualizing Data

### List of Metrics

#### etcd cluster

| Metric                                                                  | Type      | Description                                                       |
|-------------------------------------------------------------------------|-----------|-------------------------------------------------------------------|
| etcd_debugging_mvcc_db_compaction_pause_duration_milliseconds_bucket    | histogram | bucketed histogram of db compaction pause duration (ms)           |
| etcd_debugging_mvcc_db_compaction_pause_duration_milliseconds_sum       | counter   | total sum of db compaction pause duration (ms)                    |  
| etcd_debugging_mvcc_db_compaction_pause_duration_milliseconds_count     | counter   | count of db compaction pause duration (ms)                        | 
| etcd_debugging_mvcc_db_compaction_total_duration_milliseconds_bucket    | histogram | bucketed histogram of db compaction total duration (ms)           |
| etcd_debugging_mvcc_db_compaction_total_duration_milliseconds_sum       | counter   | total sum of db compaction total duration (ms)                    |  
| etcd_debugging_mvcc_db_compaction_total_duration_milliseconds_count     | counter   | count of db compaction total duration (ms)                        | 
| etcd_debugging_mvcc_db_total_size_in_bytes                              | gauge     | total size of the underlying database in bytes                    |
| etcd_debugging_mvcc_delete_total                                        | counter   | total number of deletes seen by this member                       |
| etcd_debugging_mvcc_events_total                                        | counter   | total number of events sent by this member                        |
| etcd_debugging_mvcc_index_compaction_pause_duration_milliseconds_bucket | histogram | bucketed histogram of index compaction pause duration (ms)        |
| etcd_debugging_mvcc_index_compaction_pause_duration_milliseconds_sum    | counter   | total sum of index compaction pause duration (ms)                 |
| etcd_debugging_mvcc_index_compaction_pause_duration_milliseconds_count  | counter   | count of index compaction pause duration (ms)                     |
| etcd_debugging_mvcc_keys_total                                          | gauge     | total number of keys                                              |
| etcd_debugging_mvcc_pending_events_total                                | gauge     | total number of pending events to be sent                         |
| etcd_debugging_mvcc_put_total                                           | counter   | total number of puts seen by this member                          | 
| etcd_debugging_mvcc_range_total                                         | counter   | total number of ranges seen by this member                        |
| etcd_debugging-mvcc_slow_watcher_total                                  | gauge     | total number of unsynced slow watchers                            |
| etcd_debugging_mvcc_txn_total                                           | counter   | toatl number of transactions seen by the member                   |
| etcd_debugging_mvcc_watch_stream_total                                  | gauge     | total number of watch streams                                     |
| etcd_debugging_mvcc_watcher_total                                       | gauge     | total number of watchers                                          |
| etcd_debugging_server_lease_expired_total                               | counter   | total number of expired leases                                    |
| etcd_debugging_snap_save_marshalling_duration_seconds_bucket            | histogram | the marshalling cost distributions of save called by snapshot (s) |
| etcd_debugging_snap_save_marshalling_duration_seconds_sum               | counter   | total duration of snap save marshalling (s)                       |
| etcd_debugging_snap_save_marshalling_duration_seconds_count             | counter   | count of snap save marshalling duration (s)                       |
| etcd_debugging_snap_save_total_duration_seconds_bucket                  | histogram | total latency distributions of save called by snapshot (s)        |
| etcd_debugging_snap_save_total_duration_seconds_sum                     | counter   | total sum of snapshot save duration (s)                           |
| etcd_debugging_snap_save_total_duration_seconds_count                   | counter   | count of snapshot save duration (s)                               |
| etcd_debugging_store_expires_total                                      | counter   | total number of expired store keys                                | 
| etcd_debugging_store_watch_requests_total                               | counter   | total number of incoming watch request (new or re-established)    |
| etcd_debugging_store_watchers                                           | gauge     | count of currently active watchers                                |
| etcd_disk_backend_commit_duration_seconds_bucket                        | histogram | latency distribution of commit called by backend (s)              |
| etcd_disk_backend_commit_duration_seconds_sum                           | counter   | sum of backend commit duration (s)                                |
| etcd_disk_backend_commit_duration_seconds_count                         | counter   | count of backend commit duration (s)                              |
| etcd_disk_backend_snapshot_duration_seconds_bucket                      | histogram | latency distribution of backend snapshots (s)                     |
| etcd_disk_backend_snapshot_duration_seconds_sum                         | counter   | total sum of backend snapshot duration (s)                        |
| etcd_disk_backend_snapshot_duration_seconds_count                       | counter   | count of backend snapshot duration (s)                            |
| etcd_disk_wal_fsync_duration_seconds_bucket                             | histogram | latency distribution of fsync called by wal (s)                   |
| etcd_disk_wal_fsync_duration_seconds_sum                                | counter   | total sum of fsync latency duration (s)                           |
| etcd_disk_wal_fsync_duration_seconds_count                              | counter   | count of fsync latency duration (s)                               |
| etcd_network_client_grpc_received_bytes_total                           | counter   | total numbe rof bytes received from gRPC clients                  |
| etcd_network_client_grpc_sent_bytes_total                               | counter   | total number of bytes sent to grpc clients                        |
| etcd_server_has_leader                                                  | gauge     | whether a leader exists (1 == exists, 0 == does not exist)        |
| etcd_server_leader_changes_seen_total                                   | counter   | number of leader changes seen                                     |
| etcd_server_proposals_applied_total                                     | gauge     | total number of consensus proposals applied                       |
| etcd_server_proposals_committed_total                                   | gauge     | toatl number of consensus proposals committed                     |
| etcd_server_proposals_failed_total                                      | counter   | total number of failed proposals seen                             |
| etcd_server_proposals_pending                                           | gauge     | current number of pending proposals to commit                     |

# HELP gateway_config_request_duration_seconds Bucketed histogram of request duration of Config API requests
# TYPE gateway_config_request_duration_seconds histogram
gateway_config_request_duration_seconds_bucket{le="0.0005"} 55
gateway_config_request_duration_seconds_bucket{le="0.001"} 55
gateway_config_request_duration_seconds_bucket{le="0.002"} 55
gateway_config_request_duration_seconds_bucket{le="0.004"} 55
gateway_config_request_duration_seconds_bucket{le="0.008"} 55
gateway_config_request_duration_seconds_bucket{le="0.016"} 55
gateway_config_request_duration_seconds_bucket{le="0.032"} 55
gateway_config_request_duration_seconds_bucket{le="0.064"} 55
gateway_config_request_duration_seconds_bucket{le="0.128"} 55
gateway_config_request_duration_seconds_bucket{le="0.256"} 55
gateway_config_request_duration_seconds_bucket{le="0.512"} 55
gateway_config_request_duration_seconds_bucket{le="1.024"} 55
gateway_config_request_duration_seconds_bucket{le="2.048"} 55
gateway_config_request_duration_seconds_bucket{le="4.096"} 55
gateway_config_request_duration_seconds_bucket{le="8.192"} 55
gateway_config_request_duration_seconds_bucket{le="16.384"} 55
gateway_config_request_duration_seconds_bucket{le="+Inf"} 55
gateway_config_request_duration_seconds_sum 0.0002328390000000001
gateway_config_request_duration_seconds_count 55
# HELP gateway_events_backlog Gauge of asynchronous events count waiting to be processed.
# TYPE gateway_events_backlog gauge
gateway_events_backlog 0
# HELP gateway_events_custom_processing_seconds Bucketed histogram of processing duration of an event. From receiving the asynchronous custom event to calling a function.
# TYPE gateway_events_custom_processing_seconds histogram
gateway_events_custom_processing_seconds_bucket{le="1e-05"} 0
gateway_events_custom_processing_seconds_bucket{le="2e-05"} 0
gateway_events_custom_processing_seconds_bucket{le="4e-05"} 0
gateway_events_custom_processing_seconds_bucket{le="8e-05"} 0
gateway_events_custom_processing_seconds_bucket{le="0.00016"} 0
gateway_events_custom_processing_seconds_bucket{le="0.00032"} 0
gateway_events_custom_processing_seconds_bucket{le="0.00064"} 0
gateway_events_custom_processing_seconds_bucket{le="0.00128"} 0
gateway_events_custom_processing_seconds_bucket{le="0.00256"} 0
gateway_events_custom_processing_seconds_bucket{le="0.00512"} 0
gateway_events_custom_processing_seconds_bucket{le="0.01024"} 0
gateway_events_custom_processing_seconds_bucket{le="0.02048"} 0
gateway_events_custom_processing_seconds_bucket{le="0.04096"} 0
gateway_events_custom_processing_seconds_bucket{le="0.08192"} 0
gateway_events_custom_processing_seconds_bucket{le="0.16384"} 0
gateway_events_custom_processing_seconds_bucket{le="0.32768"} 0
gateway_events_custom_processing_seconds_bucket{le="0.65536"} 0
gateway_events_custom_processing_seconds_bucket{le="1.31072"} 0
gateway_events_custom_processing_seconds_bucket{le="2.62144"} 0
gateway_events_custom_processing_seconds_bucket{le="5.24288"} 0
gateway_events_custom_processing_seconds_bucket{le="+Inf"} 0
gateway_events_custom_processing_seconds_sum 0
gateway_events_custom_processing_seconds_count 0
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
go_gc_duration_seconds{quantile="0.75"} 0
go_gc_duration_seconds{quantile="1"} 0
go_gc_duration_seconds_sum 0
go_gc_duration_seconds_count 0
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 171
# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 2.919896e+06
# HELP go_memstats_alloc_bytes_total Total number of bytes allocated, even if freed.
# TYPE go_memstats_alloc_bytes_total counter
go_memstats_alloc_bytes_total 2.919896e+06
# HELP go_memstats_buck_hash_sys_bytes Number of bytes used by the profiling bucket hash table.
# TYPE go_memstats_buck_hash_sys_bytes gauge
go_memstats_buck_hash_sys_bytes 1.444538e+06
# HELP go_memstats_frees_total Total number of frees.
# TYPE go_memstats_frees_total counter
go_memstats_frees_total 1003
# HELP go_memstats_gc_sys_bytes Number of bytes used for garbage collection system metadata.
# TYPE go_memstats_gc_sys_bytes gauge
go_memstats_gc_sys_bytes 235520
# HELP go_memstats_heap_alloc_bytes Number of heap bytes allocated and still in use.
# TYPE go_memstats_heap_alloc_bytes gauge
go_memstats_heap_alloc_bytes 2.919896e+06
# HELP go_memstats_heap_idle_bytes Number of heap bytes waiting to be used.
# TYPE go_memstats_heap_idle_bytes gauge
go_memstats_heap_idle_bytes 139264
# HELP go_memstats_heap_inuse_bytes Number of heap bytes that are in use.
# TYPE go_memstats_heap_inuse_bytes gauge
go_memstats_heap_inuse_bytes 3.923968e+06
# HELP go_memstats_heap_objects Number of allocated objects.
# TYPE go_memstats_heap_objects gauge
go_memstats_heap_objects 22136
# HELP go_memstats_heap_released_bytes_total Total number of heap bytes released to OS.
# TYPE go_memstats_heap_released_bytes_total counter
go_memstats_heap_released_bytes_total 0
# HELP go_memstats_heap_sys_bytes Number of heap bytes obtained from system.
# TYPE go_memstats_heap_sys_bytes gauge
go_memstats_heap_sys_bytes 4.063232e+06
# HELP go_memstats_last_gc_time_seconds Number of seconds since 1970 of last garbage collection.
# TYPE go_memstats_last_gc_time_seconds gauge
go_memstats_last_gc_time_seconds 0
# HELP go_memstats_lookups_total Total number of pointer lookups.
# TYPE go_memstats_lookups_total counter
go_memstats_lookups_total 132
# HELP go_memstats_mallocs_total Total number of mallocs.
# TYPE go_memstats_mallocs_total counter
go_memstats_mallocs_total 23139
# HELP go_memstats_mcache_inuse_bytes Number of bytes in use by mcache structures.
# TYPE go_memstats_mcache_inuse_bytes gauge
go_memstats_mcache_inuse_bytes 3472
# HELP go_memstats_mcache_sys_bytes Number of bytes used for mcache structures obtained from system.
# TYPE go_memstats_mcache_sys_bytes gauge
go_memstats_mcache_sys_bytes 16384
# HELP go_memstats_mspan_inuse_bytes Number of bytes in use by mspan structures.
# TYPE go_memstats_mspan_inuse_bytes gauge
go_memstats_mspan_inuse_bytes 55784
# HELP go_memstats_mspan_sys_bytes Number of bytes used for mspan structures obtained from system.
# TYPE go_memstats_mspan_sys_bytes gauge
go_memstats_mspan_sys_bytes 65536
# HELP go_memstats_next_gc_bytes Number of heap bytes when next garbage collection will take place.
# TYPE go_memstats_next_gc_bytes gauge
go_memstats_next_gc_bytes 4.473924e+06
# HELP go_memstats_other_sys_bytes Number of bytes used for other system allocations.
# TYPE go_memstats_other_sys_bytes gauge
go_memstats_other_sys_bytes 767550
# HELP go_memstats_stack_inuse_bytes Number of bytes in use by the stack allocator.
# TYPE go_memstats_stack_inuse_bytes gauge
go_memstats_stack_inuse_bytes 1.179648e+06
# HELP go_memstats_stack_sys_bytes Number of bytes obtained from system for stack allocator.
# TYPE go_memstats_stack_sys_bytes gauge
go_memstats_stack_sys_bytes 1.179648e+06
# HELP go_memstats_sys_bytes Number of bytes obtained by system. Sum of all system allocations.
# TYPE go_memstats_sys_bytes gauge
go_memstats_sys_bytes 7.772408e+06
# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE process_cpu_seconds_total counter
process_cpu_seconds_total 0.14
# HELP process_max_fds Maximum number of open file descriptors.
# TYPE process_max_fds gauge
process_max_fds 1.048576e+06
# HELP process_open_fds Number of open file descriptors.
# TYPE process_open_fds gauge
process_open_fds 9
# HELP process_resident_memory_bytes Resident memory size in bytes.
# TYPE process_resident_memory_bytes gauge
process_resident_memory_bytes 6.28736e+06
# HELP process_start_time_seconds Start time of the process since unix epoch in seconds.
# TYPE process_start_time_seconds gauge
process_start_time_seconds 1.5343639819e+09
# HELP process_virtual_memory_bytes Virtual memory size in bytes.
# TYPE process_virtual_memory_bytes gauge
process_virtual_memory_bytes 2.9646848e+07
 
