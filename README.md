# The Event Gateway

Dataflow for event-driven, serverless architectures. It routes Events (data) to Functions (serverless compute). The Event Gateway is a layer-7 proxy and realtime dataflow engine.

## Contents

1. [Philosophy](#philosophy)
2. [Motivation](#motivation)
3. [Features](#features)
   1. [Function Discovery](#function-discovery)
   2. [Pub/Sub](#pubsub)
   3. [Endpoints](#endpoints)
   4. [Multiple Emit](#multiple-emit)
   5. [Identities](#identities)
   5. [Namespaces](#namespaces)
   6. [Rules](#rules)
4. [What The Event Gateway is NOT](#what-the-event-gateway-is-not)
5. [Architecture](#architecture)
6. [HTTP API](#http-api)
7. [Plugins](#plugins)
8. [Background](#background)
9. [Comparison](#comparison)

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

- FaaS functions (AWS Lambda, Google Cloud Functions, Azure Functions, OpenWhisk Actions)
- HTTP endpoints with an HTTP method specified (e.g. GET http://example.com/function)
- group functions

#### Example: Register An AWS Lambda Function

```javascript
var sdk = require('sdk')

sdk.registerFunction("hello-world", {
  awsLambda: {
    arn: "xxx",
    region: "us-west-2",
    version: 2,
    accessKeyId: "xxx",
    secretAccessKey: "xxx"
  }
}, function(error, response) {})
```

#### Group Functions

A group function has multiple backing functions, each with a "weight" value that determines the proportional amount of traffic each should receive. If all are the same, they all receive requests equally. If one is 99 and another is 1, 99% of traffic will be sent to the one with a weight of 99.

The backing function is a function registered in the event gateway.

#### Example: Blue/Green Deployment

```javascript
var sdk = require('sdk')

// Assuming that we've already registered two functions "hello-world-v1" and "hello-world-v2"

sdk.registerFunction("hello-world-group", {
  group: {
    functions: [{
      functionId: "hello-world-v1",
      weight: 99
    }, {
      functionId: "hello-world-v2",
      weight: 1
    }]
  }
}, function(error, response) {})

// After some time we decide to route more traffic to v2 version (50%/50%)

sdk.updateFunction("hello-world-group", {
  group: {
    functions: [{
      functionId: "hello-world-v1",
      weight: 50
    }, {
      functionId: "hello-world-v2",
      weight: 50
    }]
  }
}, function(error, response) {})
```

#### Middleware

Middleware consists of functions that before or after other function and has the ability to modify the input and the output of the function. A middleware is a function registered in the event gateway. In case of error in the input middleware the error is returned and the actual function is not invoked. If the output middleware is defined the result from that output function is returned to the caller.

#### Example: Register An AWS Lambda Function With Input/Output Middleware

```javascript
var sdk = require('sdk')

sdk.registerFunction("hello-world", {
  awsLambda: {
    arn: "xxx",
    region: "us-west-2",
    version: 2,
    accessKeyId: "xxx",
    secretAccessKey: "xxx"
  },
  middleware: {
    input: "validate-data",
    output: "transform-to-html"
  }
}, function(error, response) {})
```

### Pub/Sub

Lightweight pub/sub system. Allows functions to asynchronously receive events that are published
to a topic. Functions can be configured to automatically publish their
input (e.g. useful for analyzing HTTP requests) or output to one or more topics. Instead of rewriting your
functions every time you want to send data to another place, this can be handled entirely in configuration
using the Event Gateway. This completely decouples functions from one another, reducing communication costs across
teams, eliminates effort spent redeploying functions, and allows you to easily share events across functions,
HTTP services, even different cloud providers.

* Function input or output may be configured to feed a topic
* Functions may be registered as subscribers to topics. When an event is published into the
  topic, all subscribers are called asynchronously with the event as its argument. The gateway
  recursively passes events from publishing functions to subscribing ones.
* If a function produces to two topics, and another function subscribes to both, the
  produced event is deduplicated so that the subscribing function is not called twice for
  the same thing.
* A topic may feed into another topic. This is useful for simplifying configuration of both
  publishers and subscribers who would like to operate at a variety of granularities
  and groupings. Can form graphs, rather than rigid trees that result from a hierarchical model.

#### Example: Subscribe To An Event From The Same Namespace

```javascript
var sdk = require('sdk')

sdk.createTopic("userCreated", function(error, response) {})
// Assuming that we registed "sendWelcomeEmail" function earlier
sdk.subscribeToTopic("sendWelcomeEmail", "userCreated", function(error, response) {})
```

#### Example: Subscribe To An Event From The Same Namespace Via The Framework

```yaml
gateways:
  acme:
    url: localhost

functions:
  greeter:
    handler: greeter.greeter
    events:
      - gateway.acme.userCreated
```

### Endpoints

Expose public HTTP/GraphQL/WebSocket endpoints backed by serverless functions or HTTP services.

Endpoint is a mapping between path and HTTP method, and a function.

#### Example: Create A REST API Endpoint

```javascript
var sdk = require('sdk')

// Assuming that there are following functions registered: getUser, createUser, deleteUser

sdk.createEndpoint({
  functionId: "getUser",
  method: "GET",
  path: "/users"
}, function(error, response) {})

sdk.createEndpoint({
  functionId: "createUser",
  method: "POST",
  path: "/users"
}, function(error, response) {})

sdk.createEndpoint({
  functionId: "deleteUser",
  method: "DELETE",
  path: "/users"
}, function(error, response) {})

```

The above SDK calls create a single `<The Event Gateway URL>/users` endpoint that supports three HTTP methods pointing to different backing functions.

### Multiple Emit

Optionally return multiple events, such as log messages or metrics, without sending them all back to
the caller. This plays particularly well with Pub/Sub systems. If you have an existing metrics aggregator, but don't
want to send metrics to it from within your serverless function (forcing your caller to wait while this completes)
you can return additional metrics destined for a topic of your choosing, say, "homepage-metrics". You can then
create a function that knows how to insert metrics into your existing metric system, subscribe it to "homepage-metrics", and it will forward all metrics to your existing system. You just integrated your new function with your existing systems without the function needing to know anything about them! And when you use a different metric system in the future, your code doesn't need to be updated at all. Just spin up another forwarder function, subscribe
it to the stream, and you're good to go. This is not limited to metrics!

### Identities

* Associate an ID name with a set of backing access tokens.
* Tokens are included in a header in requests to the Event Gateway.
* Priviledges may be granted and revoked by specifying a target ID, there is no need to securely transmit a secret token when granting permissions.
* No authentication requirements, we just have a mapping between a identity and their tokens.
* The identity/team behind the ID does not need to manage 50+ changing tokens for use with specific other services,
  they just keep their one token.
* The identities that we grant permissions do not have the ability to grant others the same permission by sharing tokens without asking the owner.
* Our security team is happy because we can enforce mandatory periodic key rotation. by supporting a mapping from
  ID to several tokens, we can let teams gracefully cut over to their new tokens, and phase out old ones, without
  needing a "hard cut-over" that requires downtime.
* It's not much more than a hashmap on our end, but it dramatically simplifies the lives of our identitys, and we do not
  impose any additional constraints. They can make their token the ID if they want to do fully-manual key management.

Granting and revoking rules no longer involves messy manual management of tokens. You include
your token in your requests, and you don't need to think about which token you need to use
for each backing system.

Authentication is outside the scope of the core of the system. It may be possible for plugins to implement this eventually.

#### Ownership

* all objects are assigned an owning identity upon creation, including new identities
* ownership is hierarchical

#### Identity Usage

The identity's token is hydrated into serverless.yml through an
env var or argument, based on the deployment environment.

The sdk passes along the token in an http header with every request to the API.

When it's time to perform a key rotation, a new token is added to an identity, and the owner of the identity incrementally redeploys their systems with the new
token. When the migration is complete, the old token is removed.

#### Example Identity and Namespace Usage

```
// as admin
sdk.createNamespace("analytics")
sdk.createIdentity("hendrik")
sdk.bindToken("hendrik", "120347aea9d1f25c1ca3b4d64eb561947e8418b33d")
sdk.assignNamespace("function", "f1", "analytics") // type, object, namespace
sdk.assignNamespace("identity", "hendrik", "analytics")

// hendrik can now call f1 by passing their token in a header to the gateway
```

### Namespaces

* A Namespace is a coarse-grained sandbox in which entities can interact freely.
* An identity is a member of one or more namespaces.
* Topics, Functions, and Endpoints belong to one or more namespaces.
* All actions are possible within a namespace: publishing, subscribing and calling
* All access cross-namespace is disabled by default.
* To perform cross-namespace access, use Rules (see below, not initial dev focus before Emit)

### Rules

* Probably not going to exist before Emit. Maybe never, if Namespaces are good enough. But they are likely considered below table stakes for enterprise access control.
* Rules are defined as a tuple of (Subject, Action, Object, Environment).
* Subject is an identity or function ID.
* Object may be an identity, function, endpoint, topic, or a group of them.
* Environment may relate to a logical namespace, datacenter, geographical region, the event gateway's
  conception of time, etc...
* Identities, Functions, Endpoints, and Topics may all be grouped, similar to "roles" in other systems.
* Groups, Functions, Endpoints, and Topics have an owning identity or identity group.
* The owner of a group, Function, Endpoint, or Topic may grant permissions to others on their owned resources.
* When the Event Gateway is accessed, the holder of an identity token passes the token along in a header

- functions:
  - create-function
  - delete-function
- pubsub:
  - create-topic
  - delete-topic
  - subscribe-to
  - unsubscribe-from
  - publish-to
  - stop-publishing-to
  - publish-event
- endpoints:
  - create-endpoint
  - delete-endpoint
- identities:
  - create-identity
  - delete-identity
  - bind-token-to
  - remove-token-from

#### Example Rule Usage

```
// as admin
sdk.createIdentity("alice")
sdk.bindToken("alice", "120347aea9d1f25c1ca3b4d64eb561947e8418b33d")
sdk.grant("alice", "create", "topics")
sdk.grant("alice", "create", "functions")

// as alice, who passes the token 120347aea9d1f25c1ca3b4d64eb561947e8418b33d in a header
sdk.createFunction("f1"...)
sdk.createTopic("t1"...)

// admin creates new user
sdk.createIdentity("bob")
sdk.bindToken("bob", "a9d1f25c1ca3b4d64eb561947e")
sdk.grant("bob", "create", "functions")

// alice grants permissions on things they own to bob
sdk.grant("bob", "call", "f1")
sdk.grant("bob", "subscribe", "t1")

// bob uses their new access, passing their token along
sdk.createFunction("f2"...)
sdk.subscribe("f2", "t1")

// eve does not have an identity, or they have one without permissions
sdk.grant("eve", "ownership", "t1") // FAILS
```

## Identity and Namespace Implementation Path
[See the bottom of the Access Control spec](_docs/access_control.md)

## What The Event Gateway is NOT

- it's not a replacement for message queues (no message ordering, currently weak durability guarantees only)
- it's not a replacement for streaming platforms (no processing capability and consumers group)
- it's not a replacement for existing service discovery solutions from the microservices world

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

## HTTP API

The Event Gateway exposes a RESTful configuration API.

### Function discovery

#### Register function

`POST /api/function`

Request:

- `functionId` - `string` - required, function name

Only one of the following function type can be provided.

- `awsLambda` - `object` - AWS Lambda properties:
  - `arn` - `string` - AWS ARN identifier
  - `region` - `string` - region name
  - `version` - `string` - a specific version ID
  - `accessKeyID` - `string` - AWS API key ID
  - `secretAccessKey` - `string` - AWS API key
- `gcloudFunction` - `object` - Google Cloud Function properties:
  - `name` - `string` - function name
  - `region` - `string` - region name
  - `serviceAccountKey` - `json` - Google Service Account key
- `azureFunction` - `object` - Azure Function properties:
  - `name` - `string` - function name
  - `appName` - `string` - azure app name
  - `functionsAdminKey` - `string` - Azure API key
- `openWhiskAction` - `object` - OpenWhisk Action properties:
  - `name` - `string` - action name
  - `namespace` - `string` - OpenWhisk namespace
  - `apiHost` - `string` - OpenWhisk platform endpoint, e.g. openwhisk.ng.bluemix.net
  - `auth` - `string` - OpenWhisk authentication key, e.g. xxxxxx:yyyyy
  - `apiGwAccessToken` - `string` - OpenWhisk optional API gateway access token
- `group` - `object` - Group function properties:
  - `functions` - `array` of `object` - backing functions
    - `functionId` - `string` - function ID
    - `weight` - `number` - proportion of requests destined to this function, defaulting to 1
- `http` - `object` - HTTP function properties:
  - `url` - `string` - the URL of an http or https remote endpoint

Response:

- `functionId` - `string` - function name
- `awsLambda` - `object` - AWS Lambda properties
- `gcloudFunction` - `object` - Google Cloud Function properties
- `azureFunction` - `object` - Azure Function properties
- `openWhiskAction` - `object` - OpenWhisk Action properties
- `group` - `object` - Group function properties
- `http` - `object` - HTTP function properties

#### Change configuration of group function

`PUT /api/function/<function ID>/functions`

Allows changing configuration of group function

Request:

- `functions` - `array` of `object` - backing functions
  - `functionId` - `string` - function ID
  - `weight` - `number` - proportion of requests destined to this function, defaulting to 1

Response:

- `functions` - `array` of `object` - backing functions
  - `functionId` - `string` - function ID
  - `weight` - `number` - proportion of requests destined to this function, defaulting to 1

#### Deregister function

`DELETE /api/function/<function id>`

Notes:

- used to delete all types of functions, including groups
- fails if the function ID is currently in-use by an endpoint or topic

### Endpoints

#### Create endpoint

`POST /api/endpoint`

Request:

- `functionId` - `string` - ID of backing function or function group
- `method` - `string` - HTTP method
- `path` - `string` - URL path

Response:

- `endpointId` - `string` - a short UUID that represents this endpoint mapping
- `functionId` - `string` - function ID
- `method` - `string` - HTTP method
- `path` - `string` - URL path

#### Delete endpoint

`DELETE /api/endpoint/<endpoint ID>`

#### Get endpoints

`GET /api/endpoint`

Response:

- `endpoints` - `array` of `object`
  - `endpointId` - `string` - endpoint ID, which is method + path, e.g. `GET-homepage`
  - `functionId` - `string` - function ID
  - `method` - HTTP method
  - `path` - URL path

### Pub/Sub

#### Create topic

`POST /api/topic`

Request:

- `topicId` - `string` - name of topic

Response:

- `topicId` - `string` - name of topic

#### Delete topic

`DELETE /api/topic/<topic id>`

#### Get topics

`GET /api/topic`

Response:

- `topics` - `array` of `object` - topics
  - `topicId` - `string` - topic name

#### Add subscription

`POST /api/topic/<topic id>/subscription`

Request:

- `functionId` - ID of function or function group to receive events from the topic

Response:

- `subscriptionId` - `string` - subscription ID, which is topic + function ID, e.g. `newusers-userProcessGroup`
- `topicId` - `string` - ID of topic
- `functionId` - ID of function or function group

#### Delete subscription

`DELETE /api/topic/<topic id>/subscription/<subscription id>`

#### Get subscriptions

`GET /api/topic/<topic id>/subscription`

Response:

- `subscriptions` - `array` of `object` - subscriptions
  - `subscriptionId` - `string` - subscription ID
  - `topicId` - `string` - ID of topic
  - `functionId` - ID of function or function group

#### Add publisher

`POST /api/topic/<topic id>/publisher`

Request:

- `functionId` - ID of function or function group to publish events to the topic
- `type` - either `input` or `output`

Response:

- `publisherId` - `string` - publisher ID, which is topic + function ID, e.g. `newusers-/userCreateGroup`
- `functionId` - ID of function or function group to publish events to the topic
- `type` - either `input` or `output`

#### Delete publisher

`DELETE /api/topic/<topic id>/publisher/<publisher id>`

#### Get Publishers

`GET /api/topic/<topic id>/publisher`

Response:

- `publishers` - `array` of `object` - backing functions
  - `publisherId` - `string` - publisher ID
  - `functionId` - ID of function or function group
  - `type` - either `input` or `output`

#### Publish message to the topic

`POST /api/topic/<topic id>/publish`

Request: arbitrary payload

## Plugins

Plugins are available for extending the behavior of the Event Gateway core. Examples include authentication, integration with external identity management systems,
event validation systems, etc...

They are implemented in any language as a local sidecar that adheres to the plugin calling interface. It listens on a local TCP socket for a nice mix of
interoperability and performance.

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
5. sidecar: calling function is as simple as calling method from cloud provider SDK.
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
