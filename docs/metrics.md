# Collecting and Viewing Event Gateway Metrics

This guide details how to collect and analyze metrics from the Event Gateway.

### Introduction

The Event Gateway exposes a number of [Prometheus](https://prometheus.io/)-based metrics + counters via the configuration
API to help monitor the health of the gateway.

## Contents

1. [Collecting Metrics](#collecting-metrics)
1. [Visualizing Data](#visualizing-data)
   1. [Prometheus](#prometheus)
   1. [InfluxDB](#influxdb)
1. [List of Metrics](#list-of-metrics)
   1. [Events API](#events-api)
   1. [Configuration API](#configuration-api)

### Collecting Metrics

Event Gateway metrics are exposed via a Prometheus text-based metric [exporter](https://prometheus.io/docs/instrumenting/exposition_formats/#text-format-example) and as such are
queryable by a variety of modern toolsets. While you can scrape data directly from your prometheus database, the following (non-exhaustive) list of sources can also
be configured to scrape Event Gateway metrics:

- [Telegraf](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/prometheus)
- [Datadog](https://www.datadoghq.com/blog/monitor-prometheus-metrics/)
- [AppOptics](https://docs.appoptics.com/kb/host_infrastructure/integrations/prometheus/)
- [Sensu](https://blog.sensuapp.org/the-sensu-prometheus-collector-972c441d45e)
- [Splunk](https://splunkbase.splunk.com/app/4077/#/details)
- and others ...

### Visualizing Data

Included in the [contrib](../contrib/grafana) folder are identical dashboards to visualize Event Gateway metrics with different
data sources. Depending on your setup you'll find a [Prometheus](../contrib/grafana/prometheus) and [InfluxDB](../contrib/grafana/influxdb) version
of the dashboard.

Importing a Grafana [dashboard](http://docs.grafana.org/reference/export_import/#export-and-import) is relatively straightforward provided your datasources are configured
properly within your Grafana instance. Each of our dashboards lists respective datasources as a template variable, assuming the default datasource is set
to the type for your datasource; in other words, if you import the Prometheus dashboard the template will expect your default data source to be set
to Prometheus.

If your default datasource is not either of the Prometheus or InfluxDB sources, fear not! The template variable is set as the datasource for each
panel within the dashboard, so updating the datasource is as easy as choosing another option from the template dropdown.

**NOTE:** There are two dashboards included for each datasource: an aggregated dashboard for all spaces in a given Event Gateway deployment, and a second
dashboard that can drill down to individual spaces.

#### Prometheus

You can find the source for the Prometheus dashboard [here](../contrib/grafana/prometheus).

#### InfluxDB

You can find the source for the InfluxDB dashboard [here](../contrib/grafana/influxdb).

### List of Metrics

Both Events and Configuration API exposes Prometheus metrics. The metrics are accesible via `/v1/metrics` endpoint of Configuration API. The table below outlines the specific metrics available to query from the endpoint:

#### Events API

| Metric                                          | Type      | Labels          | Description                                                                                                             |
| ----------------------------------------------- | --------- | --------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `eventgateway_events_received_total`            | counter   | `space`, `type` | total of events received                                                                                                |
| `eventgateway_events_processed_total`           | counter   | `space`, `type` | total of processed events                                                                                               |
| `eventgateway_events_dropped_total`             | counter   | `space`, `type` | total of events dropped due to insufficient processing power                                                            |
| `eventgateway_events_backlog`                   | gauge     |                 | gauge of asynchronous events count waiting to be processed                                                              |
| `eventgateway_events_custom_processing_seconds` | histogram |                 | bucketed histogram of processing duration of an event<br> (from receiving the async custom event to calling a function) |

**Labels**

- `space` - space name
- `type` - event type name

#### Configuration API

| Metric                                         | Type      | Labels                           | Description                                                   |
| ---------------------------------------------- | --------- | -------------------------------- | ------------------------------------------------------------- |
| `eventgateway_eventtypes_total`                | gauge     | `space`                          | gauge of registered event types count                         |
| `eventgateway_functions_total`                 | gauge     | `space`                          | gauge of registered functions count                           |
| `eventgateway_subscriptions_total`             | gauge     | `space`                          | gauge of created subscriptions count                          |
| `eventgateway_config_requests_total`           | counter   | `space`, `resource`, `operation` | total of Config API requests                                  |
| `eventgateway_config_request_duration_seconds` | histogram |                                  | bucketed histogram of request duration of Config API requests |

**Labels**

- `space` - space name
- `resource` - Configuration API resource, possible values: `eventtype`, `function` or `subscription`
- `operation` - Configuration API operation, possible values: `create`, `get`, `delete`, `list`, `update`
