![The Event Gateway](docs/assets/cover.png)

The Event Gateway combines both API Gateway and Pub/Sub functionality into a single event-driven experience. It's
dataflow for event-driven, serverless architectures. It routes Events (data) to Functions (serverless compute).
Everything it cares about is an event! Even calling a function. It makes it easy to share events across different
systems, teams and organizations!

Use the Event Gateway right now, by running the **[Event Gateway Example Application](https://github.com/serverless/event-gateway-example)** locally, with the _[Serverless Framework](https://serverless.com/framework/)_.

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
1.  [Motivation](#motivation)
1.  [Components](#components)
    1.  [Function Discovery](#function-discovery)
    1.  [Subscriptions](#subscriptions)
    1.  [Spaces](#spaces)
1.  [Events API](#events-api)
1.  [Configuration API](#configuration-api)
1.  [System Events](#system-events)
1.  [Plugin System](#plugin-system)
1.  [Client Libraries](#client-libraries)
1.  [Versioning](#versioning)
1.  [Comparison](#comparison)
1.  [Architecture](#architecture)
    1.  [System Overview](#system-overview)
    1.  [Clustering](#clustering)
1.  [Background](#background)

## Quick Start

### via Serverless Framework

The easiest way to get started with the Event Gateway is using the [Serverless Framework](https://serverless.com/framework/). The framework is setup to automatically download and install the Event Gateway during development of a serverless service.

Check out **[Event Gateway Example Application](https://github.com/serverless/event-gateway-example)** for a walkthrough of
using the Event Gateway locally.

### via Docker

There is a [official Docker image](https://hub.docker.com/r/serverless/event-gateway/).

```
docker run -p 4000:4000 -p 4001:4001 serverless/event-gateway -dev
```

---

If you want to install and develop with the Event Gateway without the Serverless Framework,
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
* Connectors (AWS Kinesis, AWS Kinesis Firehose)
* HTTP endpoints/Webhook (e.g. POST http://example.com/function)

Function Discovery stores information about functions allowing the Event Gateway to call them as a reaction to received
event.

#### Example: Register An AWS Lambda Function

##### curl example

```http
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

```http
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

```http
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

```http
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

```http
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
possible within a space: publishing, subscribing and invoking. All access cross-space is disabled.

Space is not about access control/authentication/authorization. It's only about isolation. It doesn't enforce any
specific subscription path.

This is how Spaces fit different needs depending on use-case:

* single user - single user uses default space for registering function and creating subscriptions.
* multiple teams/departments - different teams/departments use different spaces for isolation and for hiding internal
  implementation and architecture.

Technically speaking Space is a mandatory field ("default" by default) on Function or Subscription object that user has
to provide during function registration or subscription creation. Space is a first class concept in Config API. Config
API can register function in specific space or list all functions or subscriptions from a space.

## Events API

The Event Gateway exposes an API for emitting events. Events API can be used for emitting custom event, HTTP events and
for invoking function. By default Events API runs on `:4000` port.

### Event Definition

All data that passes through the Event Gateway is formatted as an Event, based on our default Event schema:

* `event` - `string` - the event name
* `id` - `string` - the event's instance universally unique ID (provided by the event gateway)
* `receivedAt` - `number` - the time (milliseconds) when the Event was received by the Event Gateway (provided by the event gateway)
* `data` - type depends on `dataType` - the event payload
* `dataType` - `string` - the mime type of `data` payload

Example:

```json
{
  "event": "myapp.user.created",
  "id": "66dfc31d-6844-42fd-b1a7-a489a49f65f3",
  "receivedAt": 1500897327098,
  "data": { "foo": "bar" },
  "dataType": "application/json"
}
```

When an event occurs, all subscribers are called with the event in above schema as its argument.

#### Event Data Type

The MIME type of the data block can be specified using the `Content-Type` header (by default it's
`application/octet-stream`). This allows the event gateway to understand how to deserialize the data block if it needs
to. In case of `application/json` type the event gateway passes JSON payload to the target functions. In any other case
the data block is base64 encoded.

#### HTTP Event

`http` event is a built-in type of event occurring for HTTP requests on paths defined in HTTP subscriptions. The
`data` field of an `http` event has the following structure:

* `path` - `string` - request path
* `method` - `string` - request method
* `headers` - `object` - request headers
* `host` - `string` - request host
* `query` - `object` - query parameters
* `params` - `object` - matched path parameters
* `body` - depends on `Content-Type` header - request payload

#### Invoke Event

`invoke` is a built-in event type allowing synchronous invocations. Function will react to this event only if there is a
subscription created beforehand.

### Emit a Custom Event

Creating a subscription requires `path` property (by default it's "/"). `path` indicates path under which you can push an
event.

**Endpoint**

`POST <Events API URL>/<Subscription Path>`

**Request Headers**

* `Event` - `string` - required, event name
* `Content-Type` - `MIME type string` - payload type

**Request**

arbitrary payload, subscribed function receives an event in Event schema

**Response**

Status code:

* `202 Accepted`

### Emit an HTTP Event

Creating HTTP subscription requires `method` and `path` properties. Those properties are used to listen for HTTP events.

**Endpoint**

`<method> <Events API URL>/<path>`

**Request**

arbitrary payload, subscribed function receives an event in [HTTP Event](#http-event) schema.

**Response**

Status code:

* `200 OK` with payload with function response

##### CORS

By default cross-origin resource sharing (CORS) is disabled for `http` subscriptions. It can be enabled and configured
per-subscription basis.

Event Gateway handles preflight `OPTIONS` requests for you. You don't need to setup subscription for `OPTIONS` method
because the Event Gateway will respond with all appropriate headers.

#### Path parameters

The Event Gateway allows creating HTTP subscription with parameterized paths. Every path segment prefixed with `:` is
treated as a parameter, e.g. `/users/:id`.

The Event Gateway prevents from creating subscriptions in following conflicting situations:

* registering static path when there is parameterized path registered already (`/users/:id` vs. `/users/foo`)
* registering parameterized path with different parameter name (`/users/:id` vs. `/users/:name`)

Key and value of matched parameters are passed to a function in an HTTP Event under `params` field.

##### Wildcard parameters

Special type of path parameter is wildcard parameter. It's a path segment prefixed with `*`. Wildcard parameter can only
be specified at the end of the path and will match every character till the end of the path. For examples
parameter `/users/*userpath` for request path `/users/group1/user1` will match `group1/user1` as a `userpath` parameter.

#### Respond to an HTTP Event

To respond to an HTTP event a function needs to return object with following fields:

* `statusCode` - `int` - response status code, default: 200
* `headers` - `object` - response headers
* `body` - `string` - required, response body

Currently, the event gateway supports only string responses.

### Invoking a Registered Function - Sync Function Invocation

**Endpoint**

`POST <Events API URL>/`

**Request Headers**

* `Event` - `string` - `"invoke"`
* `Function-ID` - `string` - required, ID of a function to call
* `Space` - `string` - space name, default: `default`

**Request**

arbitrary payload, invoked function receives an event in above schema, where request payload is passed as `data` field

**Response**

Status code:

* `200 OK` with payload with function response

### CORS

Events API supports CORS requests which means that any origin can emit a custom event. In case of `http` events CORS is
configured per-subscription basis.

## Configuration API

The Event Gateway exposes a RESTful JSON configuration API. By default Configuration API runs on `:4001` port.

### Function Discovery

#### Register Function

**Endpoint**

`POST <Configuration API URL>/v1/spaces/<space>/functions`

**Request**

JSON object:

* `functionId` - `string` - required, function ID
* `type` - `string` - required, provider type: `awslambda` or `http`
* `provider` - `object` - required, provider specific information about a function, depends on type:
  * for AWS Lambda:
    * `arn` - `string` - required, AWS ARN identifier
    * `region` - `string` - required, region name
    * `awsAccessKeyId` - `string` - optional, AWS API key ID. By default credentials from the [environment](http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials) are used.
    * `awsSecretAccessKey` - `string` - optional, AWS API access key. By default credentials from the [environment](http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials) are used.
    * `awsSessionToken` - `string` - optional, AWS session token
  * for HTTP function:
    * `url` - `string` - required, the URL of an http or https remote endpoint
  * for AWS Kinesis connector:
    * `streamName` - `string` - required, AWS Kinesis Stream Name
    * `region` - `string` - required, region name
    * `awsAccessKeyId` - `string` - optional, AWS API key ID. By default credentials from the [environment](http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials) are used.
    * `awsSecretAccessKey` - `string` - optional, AWS API access key. By default credentials from the [environment](http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials) are used.
    * `awsSessionToken` - `string` - optional, AWS session token
  * for AWS Firehose connector:
    * `deliveryStreamName` - `string` - required, AWS Firehose Delivery Stream Name
    * `region` - `string` - required, region name
    * `awsAccessKeyId` - `string` - optional, AWS API key ID. By default credentials from the [environment](http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials) are used.
    * `awsSecretAccessKey` - `string` - optional, AWS API access key. By default credentials from the [environment](http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials) are used.
    * `awsSessionToken` - `string` - optional, AWS session token

**Response**

Status code:

* `201 Created` on success
* `400 Bad Request` on validation error

JSON object:

* `space` - `string` - space name
* `functionId` - `string` - function ID
* `provider` - `object` - provider specific information about a function

---

#### Update Function

**Endpoint**

`PUT <Configuration API URL>/v1/spaces/<space>/functions/<function ID>`

**Request**

JSON object:

* `type` - `string` - required, provider type: `awslambda` or `http`
* `provider` - `object` - required, provider specific information about a function, depends on type:
  * for AWS Lambda:
    * `arn` - `string` - required, AWS ARN identifier
    * `region` - `string` - required, region name
    * `awsAccessKeyId` - `string` - optional, AWS API key ID
    * `awsSecretAccessKey` - `string` - optional, AWS API key
    * `awsSessionToken` - `string` - optional, AWS session token
  * for HTTP function:
    * `url` - `string` - required, the URL of an http or https remote endpoint

**Response**

Status code:

* `200 OK` on success
* `400 Bad Request` on validation error
* `404 Not Found` if function doesn't exist

JSON object:

* `space` - `string` - space name
* `functionId` - `string` - function ID
* `provider` - `object` - provider specific information about a function

---

#### Delete Function

Delete all types of functions. This operation fails if the function is currently in-use by a subscription.

**Endpoint**

`DELETE <Configuration API URL>/v1/spaces/<space>/functions/<function ID>`

**Response**

Status code:

* `204 No Content` on success
* `404 Not Found` if function doesn't exist

---

#### Get Functions

**Endpoint**

`GET <Configuration API URL>/v1/spaces/<space>/functions`

**Response**

Status code:

* `200 OK` on success

JSON object:

* `functions` - `array` of `object` - functions:
  * `space` - `string` - space name
  * `functionId` - `string` - function ID
  * `provider` - `object` - provider specific information about a function

#### Get Function

**Endpoint**

`GET <Configuration API URL>/v1/spaces/<space>/functions/<function ID>`

**Response**

Status code:

* `200 OK` on success
* `404 Not Found` if function doesn't exist

JSON object:

* `space` - `string` - space name
* `functionId` - `string` - function ID
* `provider` - `object` - provider specific information about a function

### Subscriptions

#### Create Subscription

**Endpoint**

`POST <Configuration API URL>/v1/spaces/<space>/subscriptions`

**Request**

* `event` - `string` - event name
* `functionId` - `string` - ID of function to receive events
* `path` - `string` - optional, URL path under which events (HTTP requests) are accepted, default: `/`
* `method` - `string` - required for `http` event, HTTP method that accepts requests
* `cors` - `object` - optional, in case of `http` event, By default CORS is disabled. When set to empty object CORS configuration will use default values for all fields below. Available fields:
  * `origins` - `array` of `string` - list of allowed origins. An origin may contain a wildcard (\*) to replace 0 or more characters (i.e.: http://\*.domain.com), default: `*`
  * `methods` - `array` of `string` - list of allowed methods, default: `HEAD`, `GET`, `POST`
  * `headers` - `array` of `string` - list of allowed headers, default: `Origin`, `Accept`, `Content-Type`
  * `allowCredentials` - `bool` - default: false

**Response**

Status code:

* `201 Created` on success
* `400 Bad Request` on validation error

JSON object:

* `space` - `string` - space name
* `subscriptionId` - `string` - subscription ID
* `event` - `string` - event name
* `functionId` - function ID
* `method` - `string` - optional, in case of `http` event, HTTP method that accepts requests
* `path` - `string` - optional, in case of `http` event, path that accepts requests, starts with `/`
* `cors` - `object` - optional, in case of `http` event, CORS configuration

---

#### Delete Subscription

**Endpoint**

`DELETE <Configuration API URL>/v1/spaces/<space>/subscriptions/<subscription ID>`

**Response**

Status code:

* `204 No Content` on success
* `404 Not Found` if subscription doesn't exist

---

#### Get Subscriptions

**Endpoint**

`GET <Configuration API URL>/v1/spaces/<space>/subscriptions`

**Response**

Status code:

* `200 OK` on success

JSON object:

* `subscriptions` - `array` of `object` - subscriptions
* `space` - `string` - space name
  * `subscriptionId` - `string` - subscription ID
  * `event` - `string` - event name
  * `functionId` - function ID
  * `method` - `string` - optional, in case of `http` event, HTTP method that accepts requests
  * `path` - `string` - optional, in case of `http` event, path that accepts requests
  * `cors` - `object` - optional, in case of `http` event, CORS configuration

#### Get Subscription

**Endpoint**

`GET <Configuration API URL>/v1/spaces/<space>/subscriptions/<subscription ID>`

**Response**

Status code:

* `200 OK` on success
* `404 NotFound` if subscription doesn't exist

JSON object:

* `subscriptions` - `array` of `object` - subscriptions
* `space` - `string` - space name
  * `subscriptionId` - `string` - subscription ID
  * `event` - `string` - event name
  * `functionId` - function ID
  * `method` - `string` - optional, in case of `http` event, HTTP method that accepts requests
  * `path` - `string` - optional, in case of `http` event, path that accepts requests
  * `cors` - `object` - optional, in case of `http` event, CORS configuration

### Status

Dummy endpoint (always returning `200 OK` status code) for checking if the event gateway instance is running.

**Endpoint**

`GET <Configuration API URL>/v1/status`

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

## Plugin System

The Event Gateway is built with extensibility in mind. Built-in plugin system allows reacting on system events and
manipulate how an event is processed through the Event Gateway.

_Current implementation supports plugins written only in Golang. We plan to support other languages in the future._

Plugin system is based on [go-plugin](https://github.com/hashicorp/go-plugin). A plugin needs to implement the following
interface:

```go
type Reacter interface {
	Subscriptions() []Subscription
	React(event event.Event) error
}
```

`Subscription` model indicates the event that plugin subscribes to and the subscription type. A subscription can be either
sync or async. Sync (blocking) subscription means that in case of error returned from `React` method the event won't be
further processed by the Event Gateway.

`React` method is called for every system event that plugin subscribed to.

For more details, see [the example plugin](plugin/example).

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
