![Event Gateway - React to any event with FaaS function across clouds](https://s3.amazonaws.com/assets.github.serverless/event-gateway-readme-header.png)

The Event Gateway combines both API Gateway and Pub/Sub functionality into a single event-driven experience. It's
dataflow for event-driven, serverless architectures. It routes Events (data) to Functions (serverless compute).
Everything it cares about is an event! Even calling a function. It makes it easy to share events across different
systems, teams and organizations!

Use the Event Gateway right now, by running the **[Event Gateway Getting Started Application](https://github.com/serverless/event-gateway-getting-started)** with the _[Serverless Framework](https://serverless.com/framework/)_.

Features:

* **Platform agnostic** - All your cloud services are now compatible with one another: share cross-cloud functions and events with AWS Lambda, Microsoft Azure, IBM OpenWhisk and Google Cloud Platform.
* **Send events from any cloud** - Data streams in your application become events. Centralize events from any cloud provider to get a bird’s eye view of all the data flowing through your cloud.
* **React to cross-cloud events** - You aren’t locked in to events and functions being on the same provider: Any event, on any cloud, can trigger any function. Set events and functions up like dominoes and watch them fall.
* **Expose events to your team** - Share events and functions to other parts of the application. Your teammates can find them and utilize them in their own services.
* **Extendable through middleware** - Perform data transforms, authorizations, serializations, and other custom computes straight from the Event Gateway.

The Event Gateway is a L7 proxy and realtime dataflow engine, intended for use with Functions-as-a-Service on AWS,
Azure, Google & IBM.

_The project is under heavy development. The APIs will continue to [change](#versioning) until we release a 1.0.0 version. It's not
yet ready for production applications._

[![Build Status](https://travis-ci.org/serverless/event-gateway.svg?branch=master)](https://travis-ci.org/serverless/event-gateway)

[Website](http://www.serverless.com) • [Slack](https://join.slack.com/t/serverless-contrib/shared_invite/MjI5NzY1ODM2MTc3LTE1MDM0NDIyOTUtMDgxNTcxMTcxNg) • [Newsletter](http://eepurl.com/b8dv4P) • [Forum](http://forum.serverless.com) • [Meetups](https://www.meetup.com/pro/serverless/) • [Twitter](https://twitter.com/goserverless)

## Contents

1.  [Quick Start](#quick-start)
1.  [Running the Event Gateway](#running-the-event-gateway)
1.  [Motivation](#motivation)
1.  [Components](#components)
    1.  [Function Discovery](#function-discovery)
    1.  [Subscriptions](#subscriptions)
    1.  [Spaces](#spaces)
1.  [System Events](#system-events)
1.  [APIs](#apis)
1.  [Client SDKs](#sdk)
1.  [Versioning](#versioning)
1.  [Comparison](#comparison)
1.  [Architecture](#architecture)
    1.  [System Overview](#system-overview)
    1.  [Reliability Guarantees](#reliability-guarantees)
    1.  [Clustering](#clustering)
1.  [Background](#background)

## Quick Start

### Getting Started

Looking for an example to get started? The easiest way to use the Event Gateway is with the [`serverless-event-gateway-plugin`](https://github.com/serverless/serverless-event-gateway-plugin) with the Serverless Framework.  Check out the [**Getting Started Example**](https://github.com/serverless/event-gateway-getting-started) to deploy your first service to the Event Gateway.

---

## Running the Event Gateway

### Hosted version

If you don't want to run the Event Gateway yourself, you can use the hosted version provided by the Serverless team. Please [contact us](mailto:hello@serverless.com) for an invitation to the beta.

### via Docker

There is a [official Docker image](https://hub.docker.com/r/serverless/event-gateway/).

```bash
docker run -p 4000:4000 -p 4001:4001 serverless/event-gateway -dev
```

### Binary

On macOS or Linux run the following to download the binary:

```
curl -sfL https://raw.githubusercontent.com/serverless/event-gateway/master/install.sh | sh
```

On Windows download [binary](https://github.com/serverless/event-gateway/releases).

Then run the binary in development mode with:

```bash
$ event-gateway -dev
```

---

If you want more detailed information on running and developing with the Event Gateway,
please check [Running Locally](./docs/running-locally.md) and [Developing](./docs/developing.md) guides.

## Motivation

* It is cumbersome to plug things into each other. This should be easy! Why do I need to set up a queue system to
  keep track of new user registrations or failed logins?
* Introspection is terrible. There is no performant way to emit logs and metrics from a function. How do I know
  a new piece of code is actually working? How do I feed metrics to my existing monitoring system? How do I
  plug this function into to my existing analytics system?
* Using new functions is risky without the ability to incrementally deploy them.
* The AWS API Gateway is frequently cited as a performance and cost-prohibitive factor for using AWS Lambda.

## Components

### Function Discovery

Discover and call serverless functions from anything that can reach the Event Gateway. Function Discovery supports the
following function types:

* FaaS functions (AWS Lambda, Google Cloud Functions, Azure Functions, OpenWhisk Actions)
* Connectors (AWS Kinesis, AWS Kinesis Firehose, AWS SQS)
* HTTP endpoints/Webhook (e.g. POST http://example.com/function)

Function Discovery stores information about functions allowing the Event Gateway to call them as a reaction to received
event.

#### Example: Register An AWS Lambda Function

##### curl example

```bash
curl --request POST \
  --url http://localhost:4001/v1/spaces/default/functions \
  --header 'content-type: application/json' \
  --data '{
    "functionId": "hello",
    "type": "awslambda",
    "provider":{
      "arn": "arn:aws:lambda:us-east-1:377024778620:function:bluegreen-dev-hello",
      "region": "us-east-1"
    }
}'
```

##### SDK example

```javascript
const eventGateway = new EventGateway({ url: 'http://localhost' })
eventGateway.registerFunction({
  functionId: 'sendEmail',
  type: 'awslambda',
  provider: {
    arn: 'xxx',
    region: 'us-west-2'
  }
})
```

#### Example: Function-To-Function call

##### curl example

```bash
curl --request POST \
  --url http://localhost:4000/ \
  --header 'content-type: application/json' \
  --header 'event: invoke' \
  --header 'function-id: createUser' \
  --data '{ "name": "Max" }'
```

##### SDK example

```javascript
const eventGateway = new EventGateway({ url: 'http://localhost' })
eventGateway.invoke({
  functionId: 'createUser',
  data: { name: 'Max' }
})
```

### Subscriptions

Lightweight pub/sub system. Allows functions to asynchronously receive custom events. Instead of rewriting your
functions every time you want to send data to another place, this can be handled entirely in configuration
using the Event Gateway. This completely decouples functions from one another, reducing communication costs across
teams, eliminates effort spent redeploying functions, and allows you to easily share events across functions,
HTTP services, even different cloud providers. Functions may be registered as subscribers to a custom event.
When an event occurs, all subscribers are called asynchronously with the event as its argument.

Creating a subscription requires providing ID of registered function, an event type and a path (`/` by default). The
path property indicated URL path which Events API will be listening on.

#### Example: Subscribe to an Event

##### curl example

```bash
curl --request POST \
  --url http://localhost:4001/v1/spaces/default/subscriptions \
  --header 'content-type: application/json' \
  --data '{
    "functionId": "sendEmail",
    "event": "user.created",
    "path": "/myteam"
  }'
```

##### SDK example

```javascript
const eventGateway = new EventGateway({ url: 'http://localhost' })
eventGateway.subscribe({
  event: 'user.created',
  functionId: 'sendEmail',
  path: '/myteam'
})
```

`sendEmail` function will be invoked for every `user.created` event to `<Events API>/myteam` endpoint.

#### Example: Emit an Event

##### curl example

```bash
curl --request POST \
  --url http://localhost:4000/ \
  --header 'content-type: application/json' \
  --header 'event: user.created' \
  --data '{ "name": "Max" }'
```

##### SDK example

```javascript
const eventGateway = new EventGateway({ url: 'http://localhost' })
eventGateway.emit({
  event: 'user.created',
  data: { name: 'Max' }
})
```

#### Sync subscriptions via HTTP event

Custom event subscriptions are asynchronous. There is a special `http` event type for creating synchronous
subscriptions. `http` event is an HTTP request received to specified path and for specified HTTP method. There can be
only one `http` subscription for the same `method` and `path` pair.

#### Example: Subscribe to an "http" Event

##### curl example

```bash
curl --request POST \
  --url http://localhost:4001/v1/spaces/default/subscriptions \
  --header 'content-type: application/json' \
  --data '{
    "functionId": "listUsers",
    "event": "http",
    "method": "GET",
    "path": "/users"
  }'
```

##### SDK example

```javascript
const eventGateway = new EventGateway({ url: 'http://localhost' })
eventGateway.subscribe({
  functionId: 'listUsers',
  event: 'http',
  method: 'GET',
  path: '/users'
})
```

`listUsers` function will be invoked for every HTTP GET request to `<Events API>/users` endpoint.

### Spaces

One additional concept in the Event Gateway are Spaces. Spaces provide isolation between resources. Space is a
coarse-grained sandbox in which entities (Functions and Subscriptions) can interact freely. All actions are
possible within a space: publishing, subscribing and invoking.

Space is not about access control/authentication/authorization. It's only about isolation. It doesn't enforce any
specific subscription path.

This is how Spaces fit different needs depending on use-case:

* single user - single user uses default space for registering function and creating subscriptions.
* multiple teams/departments - different teams/departments use different spaces for isolation and for hiding internal
  implementation and architecture.

Technically speaking Space is a mandatory field ("default" by default) on Function or Subscription object that user has
to provide during function registration or subscription creation. Space is a first class concept in Config API. Config
API can register function in specific space or list all functions or subscriptions from a space.

## System Events

System Events are special type of events emitted by the Event Gateway instance. They are emitted on each stage of event
processing flow starting from receiving event to function invocation end. Those events are:

* `gateway.event.received` - the event is emitted when an event was received by Events API. Data fields:
  * `event` - event payload
  * `path` - Events API path
  * `headers` - HTTP request headers
* `gateway.function.invoking` - the event emitted before invoking a function. Data fields:
  * `event` - event payload
  * `functionId` - registered function ID
* `gateway.function.invoked` - the event emitted after successful function invocation. Data fields:
  * `event` - event payload
  * `functionId` - registered function ID
  * `result` - function response
* `gateway.function.invocationFailed` - the event emitted after failed function invocation. Data fields:
  * `event` - event payload
  * `functionId` - registered function ID
  * `error` - invocation error

## APIs

[API reference](./docs/api.md)

The Event Gateway has two APIs: the Configuration API for registering functions and subscriptions, and the runtime Events API for sending events into the Event Gateway.

## SDK

* [SDK for Node.js](https://github.com/serverless/event-gateway-sdk)

## Versioning

This project uses [Semantic Versioning 2.0.0](http://semver.org/spec/v2.0.0.html). We are in initial development phase
right now (v0.X.Y). The public APIs should not be considered stable. Every breaking change will be listed in the
[release changelog](https://github.com/serverless/event-gateway/releases).

## Comparison

### What The Event Gateway is NOT

* it's not a replacement for message queues (no message ordering, currently weak durability guarantees only)
* it's not a replacement for streaming platforms (no processing capability and consumers group)
* it's not a replacement for existing service discovery solutions from the microservices world

### Event Gateway vs FaaS Providers

The Event Gateway is NOT a FaaS platform. It integrates with existing FaaS providers (AWS Lambda, Google Cloud Functions,
Azure Functions, OpenWhisk Actions). The Event Gateway enables building large serverless architectures in a unified way
across different providers.

## Architecture

### System Overview

```
                                                    ┌──────────────┐
                                                    │              │
                                                    │    Client    │
                                                    │              │
                                                    └──────────────┘
                                                            ▲
                                                            │
                                                          Event
                                                            │
                                                            ▼
                              ┌───────────────────────────────────────────────────────────┐
                              │                                                           │
                              │                   Event Gateway Cluster                   │
                              │                                                           │
                              └───────────────────────────────────────────────────────────┘
                                                            ▲
                                                            │
                                                            │
                              ┌─────────────────────────────┼─────────────────────────────┐
                              │                             │                             │
                              │                             │                             │
                              ▼                             ▼                             ▼
                      ┌───────────────┐             ┌───────────────┐             ┌───────────────┐
                      │  AWS Lambda   │             │ Google Cloud  │             │Azure Function │
                      │   Function    │             │   Function    │             │               │
                      │               │             │               │             │    Region:    │
                      │    Region:    │             │    Region:    │             │    West US    │
                      │   us-east-1   │             │  us-central1  │             │               │
                      └───────────────┘             └───────────────┘             └───────────────┘
```

### Clustering

The Event Gateway instances use a strongly consistent, subscribable DB (initially [etcd](https://coreos.com/etcd),
with support for Consul, and Zookeeper planned) to store and broadcast configuration. The instances locally
cache configuration used to drive low-latency event routing. The instance local cache is built asynchronously based on
events from backing DB.

The Event Gateway is a horizontally scalable system. It can be scaled by adding instances to the cluster. A cluster is
a group of instances sharing the same database. A cluster can be created in one cloud region, across multiple regions,
across multiple cloud provider or even in both cloud and on-premise data centers.

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

## Reliability Guarantees

### Events are not durable

The event received by Event Gateway is stored only in memory, it's not persisted to disk before processing. This means that in case of hardware failure or software crash the event may not be delivered to the subscriber. For a synchronous subscription (`http` or `invoke` event) it can manifest as error message returned to the requester. For asynchronous custom event with multiple subscribers it means that the event may not be delivered to all of the subscribers.

### Events are delivered _at most once_

Event Gateway attempts delivery fulfillment for an event only once and consequently any event received successfully by the Event Gateway is guaranteed to be received by the subscriber _at most once_. That said, the nature of Event Gateway provider implementation could result in retries under specific circumstances, but these should not cause delivering the same event multiple times. For example, Providers for AWS Services that use the AWS SDK are subject to auto retry logic that's built into the SDK ([AWS documentation on API retries](https://docs.aws.amazon.com/general/latest/gr/api-retries.html)).

AWS Lambda provider uses `RequestResponse` invocation type which means that retry logic for asynchronous AWS events doesn't apply here. Among others it means, that failed deliveries of custom events are not sent to DLQ. Please find more information in [Understanding Retry Behavior](https://docs.aws.amazon.com/lambda/latest/dg/retries-on-errors.html), "Synchronous invocation" section.

## Background

SOA came along with a new set of challenges. In monolithic architectures, it was simple to call a built-in library or
rarely-changing external service. In SOA it involves much more network communication which [is not reliable](https://en.wikipedia.org/wiki/Fallacies_of_distributed_computing). The main problems to solve include:

1.  Where is the service deployed? How many instances are there? Which instance is the closest to me? (service discovery)
2.  Requests to the service should be balanced between all service instances (load balancing)
3.  If a remote service call failed I want to retry it (retries)
4.  If the service instance failed I want to stop sending requests there (circuit breaking)
5.  Services are written in multiple languages, I want to communicate between them using the best language for the particular task (sidecar)
6.  Calling remote service should not require setting up new connection every time as it increases request time (persistent connections)

The following systems are solutions those problems:

* [Linkerd](https://linkerd.io/)
* [Istio](https://istio.io/)
* [Hystrix](https://github.com/Netflix/Hystrix/wiki) (library, not sidecar)
* [Finagle](https://twitter.github.io/finagle/) (library, not sidecar)

The main goal of those tools is to manage the inconveniences of network communication.

### Microservices Challenges & FaaS

The greatest benefit of serverless/FaaS is that it solves almost all of above problems:

1.  service discovery: I don't care! I have a function name, that's all I need.
2.  load balancing: I don't care! I know that there will be a function to handle my request (blue/green deployments still an issue though)
3.  retries: It's highly unusual that my request will not proceed as function instances are ephemeral and failing function is immediately replaced with a new instance. If it happens I can easily send another request. In case of failure, it's easy to understand what is the cause.
4.  circuit breaking: Functions are ephemeral and auto-scaled, low possibility of flooding/DoS & [cascading failures](https://landing.google.com/sre/book/chapters/addressing-cascading-failures.html).
5.  sidecar: calling function is as simple as calling method from cloud provider SDK.
6.  in FaaS setting up persistent connection between two functions defeats the purpose as functions instances are ephemeral.

Tools like Envoy/Linkerd solve different domain of technical problems that doesn't occur in serverless space. They have a lot of features that are unnecessary in the context of serverless computing.

### Service Discovery in FaaS = Function Discovery

Service discovery problems may be relevant to serverless architectures, especially when we have a multi-cloud setup or we want to call a serverless function from a legacy system (microservices, etc...). There is a need for some proxy that will know where the function is actually deployed and have retry logic built-in. Mapping from function name to serverless function calling metadata is a different problem from tracking the availability of a changing number of service instances. That's why there is a room for new tools that solves **function discovery** problem rather than the service discovery problem. Those problems are fundamentally different.

## Community

* [Slack](https://join.slack.com/t/serverless-contrib/shared_invite/MjI5NzY1ODM2MTc3LTE1MDM0NDIyOTUtMDgxNTcxMTcxNg)
* [Newsletter](http://eepurl.com/b8dv4P)
* [Forum](http://forum.serverless.com)
* [Meetups](https://www.meetup.com/pro/serverless/)
* [Twitter](https://twitter.com/goserverless)
* [Facebook](https://www.facebook.com/serverless)
* [Contact Us](mailto:hello@serverless.com)
