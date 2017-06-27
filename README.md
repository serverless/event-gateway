# The Event Gateway

Dataflow for event-driven, serverless architectures. It routes Events (data) to Functions (serverless compute). The Event Gateway is a layer-7 proxy and realtime dataflow engine.

## TOC

1. [Philosophy](#philosophy)
2. [Motivation](#motivation)
3. [Features](#features)
   1. [Function discovery](#function-discovery)
   2. [Pub/Sub](#pubsub)
   3. [Endpoints](#endpoints)
   4. [Multiple Emit](#multiple-emit)
   5. [Namespaces](#namespaces)
   6. [Access Control List](#access-control-list)
4. [What The Event Gateway is NOT](#what-the-event-gateway-is-not)
5. [Architecture](#architecture)
6. [HTTP API](#http-api)
7. [Background](#background)
8. [Comparison](#comparison)



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

Discover and call serverless functions from anything that can reach the Event Gateway. Function Discovery supports following function types:

- FaaS function (AWS Lambda, Google Cloud Function, Azure Function, OpenWhisk Action)
- HTTP endpoint with HTTP method specified (e.g. GET http://example.com/function)
- group function

#### Example: register a AWS Lambda function

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

#### Group function

A group function has multiple backing functions, each with a "weight" value that determines the proportional amount of traffic each should receive. If all are the same, they all receive requests equally. If one is 99 and another is 1, 99% of traffic will be sent to the one with a weight of 99.

Backing function is a function registered in the discovery.

#### Example: blue/grean deployent

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

### Pub/Sub

Lightweight pub/sub system. Allows functions to asynchronously receive events that are published
to a topic. Functions can be configured to automatically publish their
input (e.g. useful for analyzing HTTP requests) or output to one or more topics. Instead of rewriting your
functions every time you want to send data to another place, this can be handled entirely in configuration
using the Event Gateway. This completely decouples functions from one another, reducing communication costs across
teams, eliminates effort spent redeploying functions, and allows you to easily share events across functions,
HTTP services, even different cloud providers.

#### Example: subscribe to an event from the same namespace

```javascript
var sdk = require('sdk')

sdk.createTopic("userCreated", function(error, response) {})
// Assuming that we registed "sendWelcomeEmail" function earlier
sdk.subscribeToTopic("sendWelcomeEmail", "userCreated", function(error, response) {})
```

#### Example: subscribe to an event from the same namespace via the Framework

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

### Multiple Emit

Optionally return multiple events, such as log messages or metrics, without sending it all back to
the caller. This plays particularly well with Pub/Sub systems. If you have an existing metrics aggregator, but don't
want to send metrics to it from within your serverless function (forcing your caller to wait while this completes)
you can return additional metrics destined for a topic of your choosing, say, "homepage-metrics". You can then
create a function that knows how to insert metrics into your existing metric system, subscribe it to "homepage-metrics", and it will forward all metrics to your existing system. You just integrated your new function with your existing systems without the function needing to know anything about them! And when you use a different metric system in the future, your code doesn't need to be updated at all. Just spin up another forwarder function, subscribe
it to the stream, and you're good to go. This is not limited to metrics!

### Namespaces

Namespaces are a logical partitioning capability that enables one Event Gateway cluster to be used by multiple users, teams of users, or a single user with multiple applications without concern for undesired interaction. They provide a scope for names. Names of functions, topics, and endpoints need to be unique within a namespace, but not across namespaces.

Communication inside a namespace is open which means that function-to-function call or subscribing to an event from the same namespace doesn't involve authorization and authentication.

Communication across namespaces involves authorization and authentication process described in Access Control List section.

For better single user experience there is a default namespace in which all functions, topics, and endpoints are created if no namespace is specified. Also, by default authentication is disabled.

### Access Control List

ACL system can be used to control access to functions, topics, and endpoints. The ACL is based on tokens. The ACL is capability-based, relying on tokens to which fine-grained rules can be applied.

#### Tokens

Every token has an ID, description and rule set. Tokens are bound to a set of rules that control which Gateway resources/APIs the token has access to.

#### Rules

A rule describes the policy that must be enforced. Rules are defined on namespace level (wildcard namespace`*` can be used to apply rule to all namespaces). A rule can be enforced on following Event Gateway APIs:

- functions - function discovery operations:
  - create-function
  - delete-function
- pubsub - ACL system operations:
  - create-topic
  - delete-topic
  - subscribe-to
  - unsubscribe-from
  - publish-to
  - unbpulish-from
  - publish-event
- endpoints - endpoints operations:
  - create-endpoint
  - delete-endpoint
- tokens - tokens operations:
  - create-token
  - delete-token

#### Example: subscribe to an event from the another namespace

```javascript
var sdk = require('sdk')

// First, create a token that allows subscribing to events from your namespace "project A"

sdk.createToken({
  rules: {
    pubsub: [{
      namespace: "project-a",
      policy: "susbcribe-to"
      resource: "userCreated"
    }],
  }
}, function (err, token) {
  // token:
  {
    token: "xxx-xxx-xxx"
    rules: {
      pubsub: [{
        namespace: "project-a",
        policy: "susbcribe-to"
        resource: "userCreated"
      }],
    }
  }
})

// Pass the token to the function deployed in namespace "project B"
sdk.subscribeToTopic("sendWelcomeEmail", "userCreated", {
  namespace: "project-a",
  token: "xxx-xxx-xxx"
} function(error, response) {})
```

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

## Background

### SOA challenges

SOA introduced completely new domain of problems. In monolithic architectures, it was simple to call a service. In SOA it became a little more problematic as services are remote and calling a service involves network communication which [is not reliable](https://en.wikipedia.org/wiki/Fallacies_of_distributed_computing). The main problems to solve:

1. Where is the service deployed? With how many instances? Which instance if the closest to me? (service discovery)
2. Requests to the service should be balanced between all service instances (load balancing)
3. If a remote service call failed I want to retry it (retries)
4. If the service instance failed I want to stop sending requests there (circuit breaking)
5. Services are written in multiple languages, I want to communicate between them without writing dedicated libraries (sidecar)
6. Calling remote service should not require setting up new connection every time as it increases request time (persistent connections)

Following tech solves those problems:

- [Linkerd](https://linkerd.io/)
- [Istio](https://istio.io/)
- [Hystrix](https://github.com/Netflix/Hystrix/wiki) (library, not sidecar)
- [Finagle](https://twitter.github.io/finagle/) (library, not sidecar)

The main goal of those tools is to hide all inconveniences of network communication. They abstract network. They run on the same host as the service (or are included in the service), listen on localhost and then, based on knowledge about the whole system, know where to send a request. They use persistent connections between nodes running on the different host so there is no overhead related to connection setup (which is painful especially for secure connections).

### Microservices challenges & FaaS

The greatest benefit of serverless/FaaS is that it solves almost all of above problems:

1. service discovery: I don't care! I have a function name, that's all I need.
2. load balancing: I don't care! I know that there will be a function to handle my request (blue/green deployments still an issue though)
3. retries: It's highly unusual that my request will not proceed as function instances are ephemeral and failing function is immediately replaced with a new instance. If it happens I can easily send another request. In case of failure, it's easy to understand what is the cause.
4. circuit breaking: Functions are ephemeral and auto-scaled, low possibility of flooding/DoS & [cascading failures](https://landing.google.com/sre/book/chapters/addressing-cascading-failures.html).
5. sidecar: calling function is as simple as calling method from cloud provider SDK.
6. in FaaS setting up persistent connection between two functions defeats the purpose as functions instances are ephemeral.

Tools like Envoy/Linkerd solve different domain of technical problems that doesn't occur in serverless space. They have a lot of features that are simply not needed. Of course, it's possible to use them but that would be over-engineering.

### Service discovery in FaaS = Function discovery

Service discovery problem may be relevant to serverless architectures especially when we have multi-cloud setup or we want to call a serverless function from legacy system (by legacy I mean microservices architectures). There is a need for some proxy that will know where the function is actually deployed and have  retry logic built-in. It is a bit different problem (mapping function name -> function metadata) than (tracking where each instance of service is available). That's why there is a room for new tools that solves **function discovery** problem rather than service discovery problem. Those problems are fundamentally different.

## Comparison

### Gateway vs FaaS providers

Gateway is NOT FaaS providers. It doesn't allow to deploy or call a function. Gateway integrates with existing FaaS providers (AWS Lambda, Google Cloud Functions, OpenWhisk Actions). Gateway features enable building large, serverless architectures in a unified way across different providers.

### Gateway vs OpenWhisk

Apache OpenWhisk is a integrate serverless platform. OpenWhisk is built around three concepts:

- actions
- triggers
- rules

OpenWhisk actions, as mentioned above, is a FaaS platform. Triggers & Rules enable building event-driven systems. Those two concepts are similar to Gateway's Pub/Sub system. Though, there are few differences:

- OpenWhisk Rules doesn't integrate with other FaaS provider
- OpenWhisk doesn't provide fine-grained ACL system
- OpenWhisk doesn't enable exporting events outside OpenWhisk