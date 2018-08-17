# Collecting and Viewing Event Gateway Metrics

This guide details how to collect and analyze metrics from the Event Gateway.

## Contents
1. [Introduction](#introduction)
1. [Installing Prometheus](#installing-prometheus)
1. [Collecting Metrics](#collecting-metrics)
1. [Visualizing Data](#visualizing-data)
1. [List of Metrics](#list-of-metrics)
    1. [Events API](#events-api)
    1. [Configuration API](#configuration-api)

### Introduction

The Event Gateway exposes a number of [Prometheus](https://prometheus.io/)-based metrics + counters via the configuration 
API to help monitor the health of the gateway. The table below outlines the specific metrics available to query from the endpoint:

### Installing Prometheus

### Collecting Metrics

### Visualizing Data

### List of Metrics

Both Events and Configuration API exposes Prometheus metrics. The metrics are accesible via `/v1/metrics` endpoint of Configuration API.

#### Events API

| Metric                                          | Type      | Labels           | Description                                                                                                             |
|-------------------------------------------------|-----------|------------------|-------------------------------------------------------------------------------------------------------------------------|
| `eventgateway_events_received_total`            | counter   | `space`, `type`  | total of events received                                                                                                |
| `eventgateway_events_processed_total`           | counter   | `space`, `type`  | total of processed events                                                                                               |
| `eventgateway_events_dropped_total`             | counter   | `space`, `type`  | total of events dropped due to insufficient processing power                                                            |
| `eventgateway_events_backlog`                   | gauge     |                  | gauge of asynchronous events count waiting to be processed                                                              |
| `eventgateway_events_custom_processing_seconds` | histogram |                  | bucketed histogram of processing duration of an event<br> (from receiving the async custom event to calling a function) |

**Labels**

- `space` - space name
- `type` - event type name

#### Configuration API

| Metric                                         | Type      | Labels                           | Description                                                   |
|------------------------------------------------|-----------|----------------------------------|---------------------------------------------------------------|
| `eventgateway_eventtypes_total`                | gauge     | `space`                          | gauge of registered event types count                         |
| `eventgateway_functions_total`                 | gauge     | `space`                          | gauge of registered functions count                           |
| `eventgateway_subscriptions_total`             | gauge     | `space`                          | gauge of created subscriptions count                          |
| `eventgateway_config_requests_total`           | counter   | `space`, `resource`, `operation` | total of Config API requests                                  |
| `eventgateway_config_request_duration_seconds` | histogram |                                  | bucketed histogram of request duration of Config API requests |

**Labels**

- `space` - space name
- `resource` - Configuration API resource, possible values: `eventtype`, `function` or `subscription`
- `operation` - Configuration API operation, possible values: `create`, `get`, `delete`, `list`, `update`
