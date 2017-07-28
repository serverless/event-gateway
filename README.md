# The Event Gateway

[![serverless](http://public.serverless.com/badges/v3.svg)](http://www.serverless.com)
[![Build Status](https://travis-ci.com/serverless/event-gateway.svg?token=jjfmiKqqzKMQrFyUDpMP&branch=master)](https://travis-ci.com/serverless/event-gateway)

[Website](https://serverless.com) • [Docs](./_docs/) • [Newsletter](http://eepurl.com/b8dv4P) • [Slack](https://serverless-contrib.slack.com)


The Event Gateway combines both API Gateway and Pub/Sub functionality into a single event-driven experience, intended for use with Functions-as-a-Service on AWS, Azure, Google & IBM. It's dataflow for event-driven, serverless architectures. It routes Events (data) to Functions (serverless compute). The Event Gateway is a layer-7 proxy and realtime dataflow engine.

## Quick Start

### Running Locally

Download a binary file from the latest [release page](https://github.com/serverless/event-gateway/releases) and run `event-gateway` in `dev` mode:

```
event-gateway -dev
```

Alternatively, run in Docker container:

```
git clone git@github.com:serverless/event-gateway.git
cd event-gateway
docker build -t event-gateway .
docker run -p 4000:4000 -p 4001:4001 event-gateway -dev
```

### Register a Function

```
curl --request POST \
  --url http://127.0.0.1:4001/v1/functions \
  --header 'content-type: application/json' \
  --data '{"functionId": "hello", "provider":{"type": "awslambda", "arn": "<Function AWS ARN>", "region": "<Region>", "accessKeyId": "<Access Key ID>", "secretAccessKey": "<Secret Access Key>"}}'
```

### Subscribe to an Event

```
curl --request POST \
  --url http://127.0.0.1:4001/v1/subscriptions \
  --header 'content-type: application/json' \
  --data '{"functionId": "hello", "event": "pageVisited"}'
```

### Emit an Event

```
curl --request POST \
  --url http://127.0.0.1:4000/ \
  --header 'content-type: application/json' \
  --header 'event: pageVisited' \
  --data '{"foo": "bar"}'
```

## Contents

1. [Philosophy](#philosophy)
1. [Motivation](#motivation)
1. [Features](#features)
   1. [Function Discovery](#function-discovery)
   1. [Subscriptions](#subscriptions)
1. [Events API](#events-api)
1. [Configuration API](#configuration-api)
1. [Architecture](#architecture)
1. [What The Event Gateway is NOT](#what-the-event-gateway-is-not)
1. [Background](#background)
1. [Comparison](#comparison)

## Philosophy

- Everything we care about is an event! (even calling a function)
- Make it easy to share events across different systems, teams and organizations!

## Motivation

- It is cumbersome to plug things into each other. This should be easy! Why do I need to set up a queue system to
  keep track of new user registrations or failed logins?
- Introspection is terrible. There is no performant way to emit logs and metrics from a function. How do I know
  a new piece of code is actually working? How do I feed metrics to my existing monitoring system? How do I
  plug this function into to my existing analytics system?
- Using new functions is risky without the ability to incrementally deploy them.
- The AWS API Gateway is frequently cited as a performance and cost-prohibitive factor for using AWS Lambda.

## Features

### Function Discovery

Discover and call serverless functions from anything that can reach the Event Gateway. Function Discovery supports the following function types:

- FaaS functions (AWS Lambda, Google Cloud Functions)
- HTTP endpoints with an HTTP method specified (e.g. GET http://example.com/function)

#### Example: Register An AWS Lambda Function

```javascript
fdk.registerFunction("hello-world", {
  provider: {
    type: "awslambda",
    arn: "xxx",
    region: "us-west-2",
    accessKeyId: "xxx",
    secretAccessKey: "xxx"
  }
})
```

#### Example: Framework Integration

Every function that subscribes to an event from the gateway is automatically registered in the gateway.

```yaml
functions:
  greeter:
    handler: greeter.greeter
    events:
      - userCreated
```

#### Example: Function-To-Function call

```javascript
fdk.invoke("greeter", {
  name: "John"
})
```

### Subscriptions

Lightweight pub/sub system. Allows functions to asynchronously receive custom events. Instead of rewriting your
functions every time you want to send data to another place, this can be handled entirely in configuration
using the Event Gateway. This completely decouples functions from one another, reducing communication costs across
teams, eliminates effort spent redeploying functions, and allows you to easily share events across functions,
HTTP services, even different cloud providers. Functions may be registered as subscribers to a custom event.
When an event occurs, all subscribers are called asynchronously with the event as its argument.

#### Example: Subscribe to an Event

```javascript
// Assuming that we registered the "sendWelcomeEmail" function earlier

fdk.subscribe("sendWelcomeEmail", "userCreated")
```

#### Example: Subscribe to an Event via the Framework

```yaml
functions:
  sendWelcomeEmail:
    handler: emails.welcome
    events:
      - userCreated
```

#### Sync subscriptions via HTTP event

Custom event subscriptions are async. There is a special `http` event type for creating sync subscriptions. `http` event is
a HTTP request received on specified path and for specified HTTP method.

#### Example: Subscribe to an "http" Event

```javascript
fdk.subscribe("createUser", {
  event: "http",
  method: "GET",
  path: "/users"
})
```

#### Example: Create a REST API

```javascript
// Assuming that there are following functions registered: getUser, createUser, deleteUser

fdk.subscribe("getUser", {
  method: "GET",
  path: "/users"
})

fdk.subscribe("createUser", {
  method: "POST",
  path: "/users"
})

fdk.subscribe("deleteUser", {
  method: "DELETE",
  path: "/users"
})
```

The above FDK calls create a single `<The Event Gateway URL>/users` endpoint that supports three HTTP methods pointing to different backing functions.

#### Example: Subscribe to an "http" Event via the Framework

```yaml
functions:
  createUser:
    events:
      - http:
          path: /users
          method: POST
  getUser:
    events:
      - http:
          path: /users
          method: GET
```

## Events API

The Event Gateway exposes an API for emitting events. By default Events API runs on `:4000` port. Events API can be used for
emitting both custom and HTTP events.

### How We Define Events

All data that passes through the Event Gateway is formatted as an Event, based on our default Event schema:

- **Event (Required)**:  The Event name.
- **ID (Required)**: The Event's universally unique event ID. The Event Gateway provides this.
- **Received (Required)**: The time (milliseconds) when the Event was received by the Event Gateway. The Event Gateway provides this.
- **Data (Required):** All data associated with the Event should be contained in here.
- **Encoding (Required):** The encoding method of the data. (json, text, binary, etc.)

Example:

```json
{
  "event": "myapp.subscription.created",
  "id": "66dfc31d-6844-42fd-b1a7-a489a49f65f3",
  "received": 1500897327098,
  "data": "{\"foo\": \"bar\"}",
  "encoding": "json"
}
```

### Emit a Custom Event (Async Function Invocation)

`POST /` with `Event` header set to event name.

Request: arbitrary payload, subscribed function receives an event in above schema, where request payload is passed as `data` field

Response: HTTP status 202 - in case of success

### Emit a HTTP Event

Creating HTTP subscription requires `method` and `path` properties. Those properties are used to emit HTTP event.

`<method> /<path>`

Request: arbitrary payload, subscribed function receives an event in above schema. `data` field has following fields:

```json
{
  "data": {
    "headers": <request headers>,
    "query": <request query params>,
    "body": <request payload>
  }
}
```

Response: function response

### Invoking a Registered Function (Sync Function Invocation)

`POST /_invoke`

Request: arbitrary payload, subscribed function receives an event in above schema, where request payload is passed as `data` field

Response: function response

## Configuration API

The Event Gateway exposes a RESTful configuration API. By default Configuration API runs on `:4001` port.

### Function discovery

#### Register function

`POST /v1/functions`

Request:

- `functionId` - `string` - required, function name
- `provider` - `object` - required, provider specific information about a function, depends on type:
  - for AWS Lambda:
    - `type` - `string` - required, provider type: `awslambda`
    - `arn` - `string` - required, AWS ARN identifier
    - `region` - `string` - required, region name
    - `awsAccessKeyID` - `string` - optional, AWS API key ID
    - `awsSecretAccessKey` - `string` - optional, AWS API key
  - for HTTP function:
    - `type` - `string` - required, provider type: `http`
    - `url` - `string` - required, the URL of an http or https remote endpoint

Response:

- `functionId` - `string` - function name
- `provider` - `object` - provider specific information about a function

#### Update function

`POST /v1/functions/<function id>`

Request:

- `functionId` - `string` - required, function name
- `provider` - `object` - required, provider specific information about a function, depends on type:
  - for AWS Lambda:
    - `type` - `string` - required, provider type: `awslambda`
    - `arn` - `string` - required, AWS ARN identifier
    - `region` - `string` - required, region name
    - `awsAccessKeyID` - `string` - optional, AWS API key ID
    - `awsSecretAccessKey` - `string` - optional, AWS API key
  - for HTTP function:
    - `type` - `string` - required, provider type: `http`
    - `url` - `string` - required, the URL of an http or https remote endpoint

Response:

- `functionId` - `string` - function name
- `provider` - `object` - provider specific information about a function

#### Delete function

`DELETE /v1/functions/<function id>`

Notes:

- used to delete all types of functions
- fails if the function ID is currently in-use by a subscription

#### Get functions

`GET /v1/functions`

Response:

- `functions` - `array` of `object` - functions:
  - `functionId` - `string` - function name
  - `provider` - `object` - provider specific information about a function

### Subscriptions

#### Create subscription

`POST /v1/subscriptions`

Request:

- `event` - `string` - event name
- `functionId` - `string` - ID of function to receive events
- `method` - `string` - optionally, in case of `http` event, uppercase HTTP method that accepts requests
- `path` - `string` - optionally, in case of `http` event, path that accepts requests, it starts with "/"

Response:

- `subscriptionId` - `string` - subscription ID, which is event name + function ID, e.g. `newusers-userProcessGroup`
- `event` - `string` - event name
- `functionId` - ID of function
- `method` - `string` - optionally, in case of `http` event, HTTP method that accepts requests
- `path` - `string` - optionally, in case of `http` event, path that accepts requests

#### Delete subscription

`DELETE /v1/subscriptions/<subscription id>`

#### Get subscriptions

`GET /v1/subscriptions`

Response:

- `subscriptions` - `array` of `object` - subscriptions
  - `subscriptionId` - `string` - subscription ID
  - `event` - `string` - event name
  - `functionId` - ID of function
  - `method` - `string` - optionally, in case of `http` event, HTTP method that accepts requests
  - `path` - `string` - optionally, in case of `http` event, path that accepts requests

### Status

Dummy endpoint (always returning 200 status code) for checking if the event gateway instance is running.

`GET /v1/status`


## Architecture

```
                              AWS us-east-1 (main ─┐
                              region)              │
                              │   ┌─────────────┐  │
                              │   │             │  │
           ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─│─ ▶│    etcd     │◀ ┼ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
                              │   │             │  │                      │
           │                  │   └─────────────┘  │
                              │          ▲         │                      │
           │                  │          │         │
                              │          ▼         │                      │
           │                  │  ┌──────────────┐  │
                              │  │              │  │                      │
           │                  │  │   Gateway    │  │
                              │  │   instance   │  │                      │
           │                  │  │              │  │
                              │  └──────────────┘  │                      │
           │                  │          ▲         │
                              │          │         │                      │
           │                  │          ▼         │
                              │        ┌───┐       │                      │
GCloud us-c│ntral1───┐        │        │ λ ├┐      │           Azure West US────────┐
│          ▼         │        │        └┬──┘│      │           │          ▼         │
│  ┌──────────────┐  │        │         └───┘      │           │  ┌──────────────┐  │
│  │              │  │        └────────────────────┘           │  │              │  │
│  │   Gateway    │  │                                         │  │   Gateway    │  │
│  │   instance   │  │                                         │  │   instance   │  │
│  │              │  │                                         │  │              │  │
│  └──────────────┘  │                                         │  └──────────────┘  │
│          ▲         │                                         │          ▲         │
│          │         │                                         │          │         │
│          ▼         │                                         │          ▼         │
│        ┌───┐       │                                         │        ┌───┐       │
│        │ λ ├┐      │                                         │        │ λ ├┐      │
│        └┬──┘│      │                                         │        └┬──┘│      │
│         └───┘      │                                         │         └───┘      │
└────────────────────┘                                         └────────────────────┘
```

The Event Gateway instances use a strongly consistent, subscribable DB (initially etcd, with support for Consul, Zookeeper, and Dynamo planned) to store and broadcast configuration. The instances locally cache configuration used to drive low-latency event routing.

## What The Event Gateway is NOT

- it's not a replacement for message queues (no message ordering, currently weak durability guarantees only)
- it's not a replacement for streaming platforms (no processing capability and consumers group)
- it's not a replacement for existing service discovery solutions from the microservices world

## Background

### SOA challenges

SOA came along with a new set of challenges. In monolithic architectures, it was simple to call a built-in library or rarely-changing external service. In SOA it involves much more network communication which [is not reliable](https://en.wikipedia.org/wiki/Fallacies_of_distributed_computing). The main problems to solve include:

1. Where is the service deployed? How many instances are there? Which instance is the closest to me? (service discovery)
2. Requests to the service should be balanced between all service instances (load balancing)
3. If a remote service call failed I want to retry it (retries)
4. If the service instance failed I want to stop sending requests there (circuit breaking)
5. Services are written in multiple languages, I want to communicate between them using the best language for the particular task (sidecar)
6. Calling remote service should not require setting up new connection every time as it increases request time (persistent connections)

The following systems are solutions those problems:

- [Linkerd](https://linkerd.io/)
- [Istio](https://istio.io/)
- [Hystrix](https://github.com/Netflix/Hystrix/wiki) (library, not sidecar)
- [Finagle](https://twitter.github.io/finagle/) (library, not sidecar)

The main goal of those tools is to manage the inconveniences of network communication.

### Microservices challenges & FaaS

The greatest benefit of serverless/FaaS is that it solves almost all of above problems:

1. service discovery: I don't care! I have a function name, that's all I need.
2. load balancing: I don't care! I know that there will be a function to handle my request (blue/green deployments still an issue though)
3. retries: It's highly unusual that my request will not proceed as function instances are ephemeral and failing function is immediately replaced with a new instance. If it happens I can easily send another request. In case of failure, it's easy to understand what is the cause.
4. circuit breaking: Functions are ephemeral and auto-scaled, low possibility of flooding/DoS & [cascading failures](https://landing.google.com/sre/book/chapters/addressing-cascading-failures.html).
5. sidecar: calling function is as simple as calling method from cloud provider fdk.
6. in FaaS setting up persistent connection between two functions defeats the purpose as functions instances are ephemeral.

Tools like Envoy/Linkerd solve different domain of technical problems that doesn't occur in serverless space. They have a lot of features that are unnecessary in the context of serverless computing.

### Service discovery in FaaS = Function discovery

Service discovery problems may be relevant to serverless architectures, especially when we have a multi-cloud setup or we want to call a serverless function from a legacy system (microservices, etc...). There is a need for some proxy that will know where the function is actually deployed and have  retry logic built-in. Mapping from function name to serverless function calling metadata is a different problem from tracking the availability of a changing number of service instances. That's why there is a room for new tools that solves **function discovery** problem rather than the service discovery problem. Those problems are fundamentally different.

## Comparison

### Event Gateway vs FaaS providers

The Event Gateway is NOT a FaaS platform. It integrates with existing FaaS providers (AWS Lambda, Google Cloud Functions, OpenWhisk Actions). The Event Gateway enables building large serverless architectures in a unified way across different providers.

### Gateway vs OpenWhisk

Apache OpenWhisk is an integrated serverless platform. OpenWhisk is built around three concepts:

- actions
- triggers
- rules

OpenWhisk, as mentioned above, is a FaaS platform. Triggers & Rules enable building event-driven systems. Those two concepts are similar to the Event Gateway's Pub/Sub system. However, there are few differences:

- OpenWhisk Rules don't integrate with other FaaS providers
- OpenWhisk doesn't provide a fine-grained access control system
- OpenWhisk doesn't enable exporting events outside OpenWhisk

## Community

* [Website](https://serverless.com)
* [Blog](https://serverless.com/blog)
* [Example use-cases](./_docs/use-cases.md)
* [Email Updates](http://eepurl.com/b8dv4P)
* [Slack](https://serverless-contrib.slack.com)
* [Contact Us](mailto:hello@serverless.com)
