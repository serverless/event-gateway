# Prometheus Metrics

Both Events and Configuration API exposes Prometheus metrics. The metrics are accesible via `/v1/metrics` endpoint of Configuration API.

## Events API Metrics

| Metric Name                           | Description                                                   | Type    | Labels           |
| ------------------------------------- | ------------------------------------------------------------- | ------- | ---------------- |
| `eventgateway_events_received_total`  | Total of events received.                                     | Counter | `space`,  `type` |
| `eventgateway_events_processed_total` | Total of processed events.                                    | Counter | `space`, `type`  |
| `eventgateway_events_dropped_total`   | Total of events dropped due to insufficient processing power. | Counter | `space`, `type`  |
| `eventgateway_events_backlog`         | Gauge of asynchronous events count waiting to be processed.   | Gauge   |                  |
| `eventgateway_events_custom_processing_seconds` | Bucketed histogram of processing duration of an event. From receiving the asynchronous custom event to calling a function. | Histogram | |

### Labels

- `space` - space name
- `type` - event type name

## Configuration API Metrics

| Metric Name                               | Description                                                  | Type      | Labels                            |
| ----------------------------------------- | ------------------------------------------------------------ | --------- | --------------------------------- |
| `eventgateway_eventtypes_total`                | Gauge of registered event types count.                       | Gauge     | `space`                           |
| `eventgateway_functions_total`                 | Gauge of registered functions count.                         | Gauge     | `space`                           |
| `eventgateway_subscriptions_total`             | Gauge of created subscriptions count.                        | Gauge     | `space`                           |
| `eventgateway_config_requests_total`           | Total of Config API requests.                                | Counter   | `space`,  `resource`, `operation` |
| `eventgateway_config_request_duration_seconds` | Bucketed histogram of request duration of Config API requests. | Histogram |                                   |
### Labels

- `space` - space name
- `resource` - Configuration API resource, possible values: `eventtype`, `function` or `subscription`
- `operation` - Configuration API operation, possible values: `create`, `get`, `delete`, `list`, `update`