# The Event Gateway

[![serverless](http://public.serverless.com/badges/v3.svg)](http://www.serverless.com)
[![Build Status](https://travis-ci.org/serverless/event-gateway.svg?branch=master)](https://travis-ci.org/serverless/event-gateway)

[Website](http://www.serverless.com) • [Newsletter](http://eepurl.com/b8dv4P) • [Gitter](https://gitter.im/serverless/serverless) • [Forum](http://forum.serverless.com) • [Meetups](https://github.com/serverless-meetups/main) • [Twitter](https://twitter.com/goserverless)

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
1. [Client Libraries](#client-libraries)
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

##### curl example

```http
curl --request POST \
  --url http://localhost:4001/v1/functions \
  --header 'content-type: application/json' \
  --data '{
    "functionId": "hello",
    "provider":{
      "type": "awslambda",
      "arn": "arn:aws:lambda:us-east-1:377024778620:function:bluegreen-dev-helloa",
      "region": "us-east-1"
    }
}'
```

##### FDK example

```javascript
const eventGateway = fdk.eventGateway({ url: 'http://localhost' })
eventGateway.registerFunction({
  functionId: "sendEmail"
  provider: {
    type: "awslambda"
    arn: "xxx",
    region: "us-west-2",
  }
})
```

#### Example: Function-To-Function call

##### curl example

```http
curl --request POST \
  --url http://localhost:4000/ \
  --header ': ' \
  --header 'content-type: application/json' \
  --header 'event: invoke' \
  --header 'function-id: createUser' \
  --data '{ "name": "Max" }'
```

##### FDK example

```javascript
const eventGateway = fdk.eventGateway({ url: 'http://localhost' })
eventGateway.invoke({
  functionId: "createUser",
  data: { name: "Max" }
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

##### curl example

```http
curl --request POST \
  --url http://locahost:4001/v1/subscriptions \
  --header 'content-type: application/json' \
  --data '{
    "functionId": "sendEmail",
    "event": "user.created"
  }'
```

##### FDK example

```javascript
const eventGateway = fdk.eventGateway({ url: 'http://localhost' })
eventGateway.subscribe({
  event: "user.created",
  functionId: "sendEmail"
})
```

#### Example: Emit an Event

##### curl example

```http
curl --request POST \
  --url http://localhost:4000/ \
  --header 'content-type: application/json' \
  --header 'event: usercreated' \
  --data '{ "name": "Max" }'
```

##### FDK example

```javascript
const eventGateway = fdk.eventGateway({ url: 'http://localhost' })
eventGateway.emit({
  event: "userCreated",
  data: { name: "Max" }
})
```

#### Sync subscriptions via HTTP event

Custom event subscriptions are async. There is a special `http` event type for creating sync subscriptions. `http` event is
a HTTP request received on specified path and for specified HTTP method.

#### Example: Subscribe to an "http" Event

##### curl example

```http
curl --request POST \
  --url http://locahost:4001/v1/subscriptions \
  --header 'content-type: application/json' \
  --data '{
    "functionId": "listUsers",
    "event": "http",
    "method": "GET",
    "path": "/users"
  }'
```

##### FDK example

```javascript
const eventGateway = fdk.eventGateway({ url: 'http://localhost' })
eventGateway.subscribe({
  functionId: 'listUsers',
  event: 'http',
  method: 'GET',
  path: '/users'
})
```

## Client Libraries

- [FDK for Node.js](https://github.com/serverless/fdk)

## Events API

The Event Gateway exposes an API for emitting events. By default Events API runs on `:4000` port. Events API can be used for
emitting both custom and HTTP events.

### How We Define Events

All data that passes through the Event Gateway is formatted as an Event, based on our default Event schema:

- `event` - `string` - the event name
- `id` - `string` - the event's instance universally unique ID (provided by the event gateway)
- `receivedAt` - `number` - the time (milliseconds) when the Event was received by the Event Gateway (provided by the event gateway)
- `data` - type depends on `dataType` - the event payload
- `dataType` - `string` - the mime type of `data` payload

Example:

```json
{
  "event": "myapp.subscription.created",
  "id": "66dfc31d-6844-42fd-b1a7-a489a49f65f3",
  "receivedAt": 1500897327098,
  "data": {"foo": "bar"},
  "dataType": "application/json"
}
```

#### Event Data Type

The MIME type of the data block can be specified using the `Content-Type` header (by default it's `application/octet-stream`). This allows the event gateway to understand how to deserialize the data block if it needs to. In case of `application/json` type the event gateway passes JSON payload to the target functions. In any other case the data block is Base64 encoded.

### Emit a Custom Event (Async Function Invocation)

`POST /` with `Event` header set to an event name. Optionally `Content-Type: <MIME type>` header can be set to specify payload encoding.

Request: arbitrary payload, subscribed function receives an event in above schema, where request payload is passed as `data` field

Response: `202 Accepted` in case of success

### Emit an HTTP Event

Creating HTTP subscription requires `method` and `path` properties. Those properties are used to listen for HTTP events.

`<method> /<path>`

Request: arbitrary payload, subscribed function receives an event in above schema. `data` field has the following fields:

```
{
  ...
  "data": {
    "headers": <request headers>,
    "query": <request query params>,
    "body": <request payload>,
    "path": <request URL path>,
    "method": <request method>
  }
  ...
}
```

Response: function response

### Respond to an HTTP Event

To respond to an HTTP event a function needs to return object with following fields:

- `statusCode` - `int` - response status code, default: 200
- `headers` - `object` - response headers
- `body` - `string` - response body

Currently, the event gateway supports only string responses.

### Invoking a Registered Function (Sync Function Invocation)

`POST /` with `Event` header set to `invoke` and `Function-ID` set to function ID.

Request: arbitrary payload, invoked function receives an event in above schema, where request payload is passed as `data` field

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

`PUT /v1/functions/<function id>`

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
- `path` - `string` - optionally, in case of `http` event, path that accepts requests, starts with `/`

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

* [Email Updates](http://eepurl.com/b8dv4P)
* [Serverless Forum](http://forum.serverless.com)
* [Gitter Chatroom](https://gitter.im/serverless/serverless)
* [Serverless Meetups](http://www.meetup.com/serverless/)
* [Facebook](https://www.facebook.com/serverless)
* [Twitter](https://twitter.com/goserverless)
* [Contact Us](mailto:hello@serverless.com)
