# The Event Gateway

Dataflow for event-driven, serverless architectures. The Event Gateway is a layer-7 proxy and realtime dataflow engine.

## Philosophy

- Everything we care about is an event! (even calling a function)
- Make it easy to share events across different systems!

## Motivation

- It is cumbersome to plug things into each other. This should be easy! Why do I need to set up a queue system to
keep track of new user registrations or failed logins?
- Introspection is terrible. There is no performant way to emit logs and metrics from a function. How do I know
a new piece of code is actually working? How do I feed metrics to my existing monitoring system? How do I
plug this function into to my existing analytics system?
- Using new functions is risky without the ability to incrementally deploy them.
- The AWS API Gateway is frequently cited as a performance and cost-prohibitive factor for using Lambda.

## Features

### Pub/Sub

Lightweight pub/sub system. Allows functions to asynchronously receive events that are published
to a topic. Functions can be configured to automatically publish their
input (useful for analyzing HTTP requests etc...) or output to one or more topics. Instead of rewriting your
functions every time you want to send data to another place, this can be handled entirely in configuration
using the Event Gateway. This completely decouples functions from one another, reducing communication costs across
teams, eliminates effort spent redeploying functions, and allows you to easily share events across functions,
HTTP services, even different cloud providers.

### Function Discovery

Discover and call serverless functions from anything that can reach the Event Gateway.

### Endpoints

Expose public HTTP/GraphQL/REST/WebSocket endpoints backed by serverless functions or HTTP services.

### Multiple Emit

Optionally return multiple events, such as log messages or metrics, without sending it all back to
the caller. This plays particularly well with Pub/Sub systems. If you have an existing metrics aggregator, but don't
want to send metrics to it from within your serverless function (forcing your caller to wait while this completes)
you can return additional metrics destined for a topic of your choosing, say, "homepage-metrics". You can then
create a function that knows how to insert metrics into your existing metric system, subscribe it to "homepage-metrics",
and it will forward all metrics to your existing system. You just integrated your new function with your existing
systems without the function needing to know anything about them! And when you use a different metric system
in the future, your code doesn't need to be updated at all. Just spin up another forwarder function, subscribe
it to the stream, and you're good to go. This is not limited to metrics!

## What The Event Gateway is NOT:

- it's not a replacement for message queues (no message ordering, currently weak durability guarantees only)
- it's not a replacement for streaming platforms (no processing capability and consumers group)
- it's not a replacement for existing service discovery solutions from the microservices world

## Use cases

### REST API

serverless.yaml:

```yaml
gateways:
  acme:
    url: gateway.serverless.com/acme
    apikey: xxx

functions:
  greeter:
    handler: greeter.greeter
    events:
      - gateway.acme.http:
          method: GET
          path: greet

// General event naming convention
      - gateway.<gateway name>.<event name>:
          <props>
```

### SaaS webhooks - Predifined events

Gateway supports predefined set of events that correspond to webhooks exposed by companies like e.g. GitHub, Twilio. Each predefined
event has a specific structure.

serverless.yaml:

```yaml
gateways:
  acme:
    url: gateway.serverless.com/acme
    apikey: xxx

functions:
  greeter:
    handler: greeter.greeter
    events:
      - gateway.acme.github:
        repo: serverless/serverless
        type: commit_comment
```

### Reacting on custom events

serverless.yaml:

```yaml
gateways:
  acme:
    url: gateway.serverless.com/acme
    apikey: xxx

functions:
  greeter:
    handler: greeter.greeter
    events:
      - gateway.acme.userCreated
```

### Reacting on custom events from multiple gateways

serverless.yaml:

```yaml
gateways:
  acme:
    url: gateway.serverless.com/acme
    apikey: xxx
  evilcorp:
    url: gateway.serverless.com/evilcorp
  internal:
    url: 127.0.0.1

functions:
  welcomeEmail:
    handler: emails.welcomeEmail
    events:
      - gateway.internal.userCreated
    events:
      - gateway.acme.userCreated
```

### Publishing custom events

Publishing events happens in the code usind FDK

```javascript
const fdk = require('serverless/fdk')

module.exports.hello = fdk().handler((event, ctx) => {
  fdk.emit("acme", {
    "userCreated": {
      id: "xxx"
    }
  })

  return "hello"
})
```

### Publishing multiple custom events

```javascript
const stdlib = require('stdlib')

module.exports.hello = stdlib().handler((event, ctx) => {
  stdlib.emit({
    "userCreated": {
      id: "xxx"
    }
  })

  stdlib.emit({
    "userSaved": {
      id: "xxx"
    }
  })

  return "hello"
})
```

### Event discovery

Event discovery is available via Platform UI or the framework.

```
$ serverless events search "user*"

- userCreated
  gateway: acme
  published at: 3 minutes age

- userDeleted
  gateway: evilcorp
  published at: 2 days ago
```

### Reacting to operational data from function

There is special `function` event that allows reacting on function low-level events like:

- invocation - metrics about invocation (memory usage, duration)
- logs - logs reported by the function
- result - raw result returned by the function

serverless.yaml:

```yaml
gateways:
  acme:
    url: gateway.serverless.com/acme
    apikey: xxx

functions:
  pushMetricsToDataDog:
    handler: metrics.datadog
    events:
      - gateway.acme.function:
        - name: welcomeEmail
        - type: invocation
```

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

## API (for MVP)

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
* used to delete all types of functions, including groups
* fails if the function ID is currently in-use by an endpoint or topic

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
	- `id` - `string` - endpoint ID, which is method + path, e.g. `GET-/homepage`
	- `functionId` - `string` - function ID
	- `method` - HTTP method
	- `path` - URL path

### Pub/Sub

#### Create topic

`POST /api/topic`

Request:

- `id` - `string` - name of topic

Response:

- `id` - `string` - name of topic

#### Delete topic

`DELETE /api/topic/<topic id>`

#### Get topics

`GET /api/topic`

Response:

- `topics` - `array` of `object` - topics
  - `id` - `string` - topic name

#### Add subscription

`POST /api/topic/<topic id>/subscription`

Request:

- `functionId` - ID of function or function group to receive events from the topic

Response:

- `subscriptionId` - `string` - subscription ID, which is topic + function ID, e.g. `newusers-/userProcessGroup`
- `functionId` - ID of function or function group

#### Delete subscription

`DELETE /api/topic/<topic id>/subscription/<subscription id>`

#### Get subscriptions

`GET /api/topic/<topic id>/subscription`

Response:

- `subscriptions` - `array` of `object` - backing functions
  - `subscriptionId` - `string` - subscription ID
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

# Brainstorm / Specification draft

The following document includes a lot of random ideas that might not get to the implementation phase.

## Overview

```
 Internet────────────────────────────────────────────┐      Internal infrastructure (e.g. AWS)────────────────────┐
 │                                                   │      │                                                     │
 │                                                   │      │                                                     │
 │                                                   │      │                                                     │
 │                                                   │      │            ┌───┐         ┌───┐                      │
 │                                                   │      │            │ λ │         │ λ │                      │
 │                                                   │      │            └───┘         └───┘                      │
 │                                                   │      │              │             ▲                        │
 │         ┌─────────────────┐                       │      │           publish      react on                     │
 │         │                 │                       │      │           events        events                      │
 │         │   Mobile apps   │◀───HTTP & push    ┌───┴──────┴────────┐ (pub/sub)    (pub/sub)                     │
 │         │                 │   notifications   │┌───────┐          │     │             │                        │
 │         └─────────────────┘         │         ││       │          │     │             │                        │
 │                                     │         ││       │          │     │             │                        │
 │                                     │         ││ Edge  │   Gateway│     │             │    function   ┌───┐    │
 │         ┌─────────────────┐         ├────────▶││ proxy │          │─────┴─────────┬───┴────metadata ─▶│ λ │    │
 │         │                 │         │         ││       │          │               │       (discovery) └───┘    │
 │         │  Browser apps   │◀───GraphQL &      ││       │          │               │                     │      │
 │         │                 │    WebSockets     │└───────┘          │               │                     │      │
 │         └─────────────────┘                   └───┬──────┬────────┘           configure              call a    │
 │                                                   │      │                    function              function   │
 │                                                   │      │                 (config store)             (FDK)    │
 │                                                   │      │                        │                     │      │
 │                                                   │      │                        ▼                     ▼      │
 │                                                   │      │                      ┌───┐                 ┌───┐    │
 │                                                   │      │                      │ λ │                 │ λ │    │
 │                                                   │      │                      └───┘                 └───┘    │
 │                                                   │      │                                                     │
 │                                                   │      │                                                     │
 └───────────────────────────────────────────────────┘      └─────────────────────────────────────────────────────┘
```

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

## Motivation

- enable developers to build FaaS backends for modern web applications by providing event-based communication layer
- enable developers to build micro-services (serverless) architectures by providing discovery features and communication layer
- enable developers to build event-driven backend systems by providing configuration store and communication layer

## Concepts

### Event discovery

Event discovery is a sub service for storing metadata of events that are handled by the Gateway. Event discovery is a single source of truth about types of events that occur in the system. Events can be grouped into streams by prefixing name with stream name (separated with "/") e.g. "users/userCreated", "cart/itemAdded".

#### Schema

Event schema can be defined explicitly using Gateway API or implicitly with the first occurrence of the event. Schema created implicitly can be changed. After schema creation every event with the same name has to be valid against schema.

#### Access types

There are two types of access:
- private (default) - event occurs in the backend system, internally, only functions from Function discovery can subscribe to them.
- public - events that can be published to external subscribers (browser application via WebSockets, pushed to a mobile app via SNS, or pushed to external HTTP endpoint)

### Pub/Sub

Competitors: Firebase Cloud Messaging, RabbitMQ, AWS SNS

Pub/Sub sub system allows publishing and subscribing to custom events. Pub/Sub sub system is lightweight message broker tailored for FaaS architectures.

#### Publication

An event can be published via Gateway API. If there is no schema registered in Events discovery for this type of event, new schema is created ad-hoc. If the event is not valid against existing schema error is returned and the event is not published. Events can also be pulled from other systems (like Kafka, RabbitMQ) via connectors that forward events to the Gateway.

#### Subscription

A subscriber can subscribe to a specific event or group of events. Subscriptions are created using Gateway API by providing event/stream name and subscription target. Target types:

- function - target function has to be registered in Function discovery before creating the subscription (for private & public events).
- HTTP endpoint - any HTTP endpoint (for public events)
- WebSockets channel (for public events, more about WebSockets channels in Edge proxy section)
- any external source via connectors (for private & public events)

##### Connectors

A connector is a Gateway plugin that imports/exports events from/to external sources like Kafka, Kinesis or RabbitMQ. Events are either pulled or pushed by a connector. A connector has a full access (including private events) to events occurring in the system. It may be helpful for importing events from existing legacy system or for archiving/event sourcing purpose.

```
                        ┌─────────────────────┐                  ┌─────┐
                   ┌────┴────┐                │     event: ─────▶│  λ  │
┌────────────┐     │  Kafka  │                │   userCreated    └─────┘
│   Kafka    │◀───▶│connector│                │        │
└────────────┘     └────┬────┘  Gateway       │        │
                   ┌────┴────┐                │────────┤
┌────────────┐     │RabbitMQ │                │        │         ┌─────┐
│  RabbitMQ  │◀───▶│connector│                │     event: ─────▶│  λ  │
└────────────┘     └────┬────┘                │   userVisited    └─────┘
                        └─────────────────────┘
```

A connector can be either import, export or both. Example of archiving all events to S3 with S3 connector configured as a export connector:

```
                      ┌─────────────────────────────────┐
                      │             Gateway             │
                      │                                 │
                      │             ┌───┐               │
                      │      ┌─────▶│ λ │               │
┌─────────────────────┴┐     │      └───┘              ┌┴──────────────────┐
│  Kinesis connector   │     │                         │   S3 connector    │
│(configured to import │     │                         │  (configured to   │
│ events from specific │─────┴────────────────────────▶│ export all events │
│   Kinesis stream)    │                               │   to S3 bucket)   │
└─────────────────────┬┘                               └┬──────────────────┘
                      │                                 │
                      └─────────────────────────────────┘
```

#### Pub/Sub & Serverless framework

In case of Serverless Framework function can subscribe using event sources:

```
gateway:
  token: xxx
  url: gateway.acme.com

functions:
  hello:
    handler: index.run
    events:
      - gateway:
          event: userCreated
      - gateway:
          event: userDeleted
```


### Function discovery

Competitors: Consul, etcd

Function discovery is a low latency, highly available sub service storing metadata and health info about registered function. Function discovery maps functions name to function metadata (provider, provider specific ID, region, timeout, etc.). Function discovery is a source of data for FDK enabling calling functions from other functions or legacy systems.

Functions can be grouped into services/apps by prefixing name with service/app name (separated with "/") e.g. "users/create", "cart/addItem".

#### Problems to solve

- calling a function without knowing where the function is deployed (provider, region)
- calling a function that exposes HTTP endpoint (e.g. HTTP endpoint via APIG or now.sh/stdlib/clay functions) without knowing exact URL
- insight into which function are working properly, which are failing (because of timeout, throttle, or runtime error).

#### Registration

Function can be registered via Gateway API. Function types:

- FaaS function:
  - name
  - instances:
    - provider
    - provider specific ID
    - region
- function exposing HTTP
  - name
  - HTTP method
  - instances:
    - region
    - url

#### Discovery

A function can be discovered with a function name. Function discovery returns function metadata that allows calling the function.

#### Health status

Function discovery stores information about functions health. Health information is advisory and doesn't affect function metadata returned from Discovery. The function is healthy if all calls are successfully handled. The unhealthy function is a function that couldn't handle request because of timeout, throttle or runtime error.

Gateway exposes API to provide info about function health (failed calls).

#### Access types

There are two types of access:
- private (default) - functions that can be called from within the same infrastructure/system
- public - functions that can be called via edge proxy endpoints. More about public functions in "Edge proxy" section.

#### Multi-regional deployments

It's NOT mandatory to provide region during registration. It's additional data that distinguish different instances of the same function enabling users to provide low latency, reliable experience for end users.

#### FDK & function discovery

FDK uses Gateway client internally. Flow:

- `fdk.request()`, `fdk.call()` or `fdk.trigger()` call
- FDK fetches function metadata from the Gateway via Gateway client
- If function is deployed to multiple regions return metadata of the closest function instance
- FDK using ARN/URL calls the function using provider SDK (e.g. `aws-sdk`) or makes HTTP requests
- notify Gateway service about any issues during the remote call (e.g. function not found, timeout, throttled, HTTP endpoint returns an error)

### Config store

Competitors: Consul, etcd

Config store is a low latency, highly available, simple DB for storing key (string) -> value (string) pairs. It can be used to store any kind of data. The intention is to replace environment variables as a way to pass configuration values.

Values can be grouped into folders by prefixing key with folder name e.g. `users/apikey`.

#### Problems to solve

- storing configuration that can be dynamically changed without requiring function redeploy
- providing a secure way for storing credentials used by functions

#### FDK & config store

FDK provides simple say for fetching and storing data in config store.

```
fdk.set('users/twilioKey', 'xxx')
fdk.get('users/twilioKey')
```

#### Implications of Storage Semantics

For configuration, this system has this desired behavior:

1. something wants to configure something
1. everything that needs to know about that configuration reacts to it

We benefit from strongly consistent atomic writes because we want all configuration to
reach everything that needs to be configured. Weakly consistent systems
may not observe all updates, and we will end up with broken or half-working
features. We also benefit from efficient notification primitives for
keeping the overall system performant, such as watches. Without watches,
we have to continuously scan ranges of keys in the database until we
find new data, which is enormously expensive, slow, and hard to scale.

Applied to security, where we want to revoke a user's privileges completely,
if we are using a database without the ability to perform atomic writes,
even if it is strongly consistent,
then one component may try updating a user's account by first reading
their data, modifying it locally, then trying to update it in the database.
If we don't have atomic updates, the ordering could end up being this:

```
node A reads user Bob's data
node A locally updates Bob's data to add admin privileges
node B deletes Bob's data to remove all privileges
node B receives "write successful" from database
node A writes Bob's new data to the system with admin privileges
Node B then tells the system that was trying to delete Bob "Bob sucessfully deleted"
Bob steals all of the company's secrets
```

This is not desirable. However, some strongly consistent databases let us
perform compare-and-swap operations that let us atomically update Bob, without
losing any intermediate updates.

```
node A reads user Bob's data
node A locally updates Bob's data to add admin privileges
node B deletes Bob's data to remove all privileges
node B receives "write successful" from database
node A tries to do "update Bob unless changed since read" which fails
Node B then tells the system that was trying to delete Bob "Bob sucessfully deleted"
Bob is locked out and steals nothing
```

Atomic updates generally significantly reduce the cognitive burden for
building a system which must store and react to state changes in different
components. Watches are another significant help for this, as they
notify interested systems in relevant changes.

Watches are how we apply the event-driven model to our database.
When interesting changes happen, interested parties react to them.
This decouples the emitter from the reactor, greatly simplifying
interactions in the system. Without watches, we need to
have some way of detecting changes in a database. If we can scan through
all keys, we can do an O(N) traversal of the entire database, which
does not scale very high, but could work alright. If we don't have
the ability to scan through all data, we may actually have no way of
learning what changes are unless there is a top-level key that is set
that holds everything. This does not scale beyond a couple kilobytes.

#### Storage Options

##### stateless gateway services backed by zk/etcd/consul cluster, abstracted by docker/libkv

pros

* flexible support for the three most popular configuration databases
* very clear operational characteristics, does not confuse anyone about what's happening
* the gateway is fully stateless, easy to autoscale, clear semantics for operators
* lowest amount of work for us
* write once, run anywhere in the same way
* allows users to easily take advantage of existing database skills, tools, backup tools, monitoring, etc...

cons

* requires users to run their own cluster (they are already running things though, so this isn't a high marginal cost)

##### embedded etcd in gateway, cluster of 3 or 5 active as "leaders", rest of gateways are stateless

pros

* single binary

cons

* unclear operational characteristics, when the cluster gets wedged it may be extremely hard to debug
* harder than running your own cluster, because you can't reuse existing database skills
* very hard on operators when things go wrong
* harder on operators to get things safely set up
* still effectively have a separate cluster if you want to autoscale without accidentally losing leaders
* creating a reliable "autopilot" etcd deployment system took tyler 4 months in the past

##### embedded etcd in gateway, single node configured as "leader", rest of gateways are stateless

pros

* easy to set up
* can be used in combination with a separate etcd cluster for the best of both worlds
* does not require setting up a cluster to try out, demo, or run in small deployments
* clear operational characteristics, everyone knows it's not reliable

cons

* single point of failure on the single leader node

##### dynamo + spanner + cosmosdb + on prem other databases

pros

* single binary
* easy deployment for users on cloud providers
* one fewer piece to keep running

cons

* dramatically increases the complexity for building the system for us
* we need to target multiple consistency models (hard+++)
* need to target systems without watch semantics (hard+++)
* need to target systems without atomic write semantics (hard+++)
* we need to spend much more effort on testing
* we need to spend much more effort on fixing bugs
* we need to spend much more effort on writing monitoring code
* we need to master all of these databases in order to program against them

##### eventually consistent gossip-based config sharing with CRDTs

pros

* single binary
* no external dependencies, even on databases

cons

* if all instances go away, the configuration is gone
* we're basically building our own distributed database (hard+++++)
* need to create backup tooling
* unclear operational characteristics for people
* huge extra effort needs to go into correctness testing

##### just config files, reloaded on file change

pros

* single binary
* flexible
* can be used with any of the other approaches to decouple functionality

cons

* users don't get an API for the gateway

#### Recommendation

Keep the system stateless to make it easy to operate. Store state
in a database that supports atomic updates and watches, such as
etcd, zookeeper, or consul. Use the docker/libkv library to
support all 3 backing databases. This lets us treat every
deployment environment the same way, regardless of whether
it's in a cloud provider or on-prem. This combination of
operational clarity and database semantics will significantly
lower the amount of effort we need to put into engineering over time.

Purely for demo and trial purposes, allow a gateway to be started
with an `--embedded-master` flag which will start an embedded etcd instance
that other gateways can use as a shared backing store. This allows people to
try out the system without standing up an etcd cluster first.

The main downside, having to run a separate cluster in production, is probably
not a big deal for people who are interested in running this themselves.

Assumptions relating to this cost, and it not seeming very high:

1. most users will fall into these camps:
  * small/medium orgs interested in trying out locally but using the SaaS offering
  * orgs with more significant engineering resources who already run zk/etcd/consul
  * orgs with more significant engineering resources who are not averse to running zk/etcd/consul
  * orgs who are comfortable paying compose.com $30/mo for hosted etcd
1. users who are not interested in paying for SaaS but still want to run the gateway
   themselves WITHOUT running a database are unlikely to be very upset when they occasionally
   need to reconfigure the gateway when the single master goes away.

### ACL system

*ACL system is highly inspired by [Consul's ACL system](https://www.consul.io/docs/guides/acl.html) and AWS IAM.*

ACL system can be used to control access to functions and events. The ACL is based on tokens. The ACL is capability-based, relying on tokens to which fine-grained rules can be applied.

#### Tokens

Every token has an ID, description and rule set. Tokens are bound to a set of rules that control which Gateway resources/APIs the token has access to.

#### Rules

A rule describes the policy that must be enforced. Rules are prefix-based, allowing operators to define different namespaces (config store folder, service, stream). A rule can be enforced on following Gateway APIs:

- config - config store operations
- function - function discovery operations
- event - event discovery operations
- acl - ACL system operations
- endpoint - edge proxy endpoints operations

Rules can make use of one or more policies. Policies can have following dispositions:

- read - allow the resource to be read but not modified
- write - allow the resource to be read and modified
- deny - do not allow the resource to be read or modified

With prefix-based rules, the most specific prefix match determines the action. This allows for flexible rules like an empty prefix to allow read-only access to all resources, along with some specific prefixes that allow write access or that are denied all access.

**Example**

Following rule set gives permissions for writing values in config store under `users/` folder and allows function react on `userCreated` event:

```json
{
  "id": "94f7efe8-db7d-4123-8d84-b9c75eaa495d",
  "description": "users service functions token",
  "rules": {
    "config": [{
      "resource": "users/",
      "policy": "write"
    }],
    "event": [{
      "resource": "userCreated",
      "policy": "read"
    }]
  }
}
```

### Edge proxy

Competitors: AWS API Gateway

Edge proxy (EP) exposes endpoints that can be accessed publicly (on the Internet) or internally. Endpoints types:

- HTTP endpoint - simple HTTP endpoint that allows calling functions registered in Function discovery. Similar to AWS API Gateway.
- GraphQL endpoint - this endpoint allows exposing multiple functions via GraphQL endpoint without a need to create GraphQL server. EP acts as a GraphQL server and takes care of calling backend functions. Developer is only responsible for providing GraphQL schema.
- WebSockets channels - a bridge between Pub/Sub and web browser (or any WebSockets-compatible client)
- SNS/Firebase Cloud Messaging/etc. - a bridge between Pub/Sub and mobile devices or other supported targets

Gateway exposes API for creating/deleting endpoints.

#### HTTP & GraphQL endpoints

Those endpoints accept HTTP request and forward them to backend functions (prior registering them in Function discovery).

#### WebSockets channels

WebSockets channels endpoints enable two-way communication between backend function and the browser app. Browser app can subscribe to public events defined in Event discovery. It can also publish a public event to the Pub/Sub system. Gateway takes care of authorization and authentication of WebSockets connection.

#### SNS/Firebase Cloud Messaging/etc.

Gateway product can push messages to mobile devices or other custom targets via existing cloud services. The difference between Pub/Sub connectors and endpoints is that connectors have low-level access to all events handled by Gateway (both public and private events). Endpoints are only means of transport for events pushed to the devices.

## Use Cases

### Modern web application

- frontend application running in a browser or mobile app,
- backend accessible via GraphQL or REST API,
- WebSockets support for reactivity

### Microservices

- many small functions
- functions deployed on different providers and available via different protocols (AWS Lambda, HTTP, GRPC) (functions discovery)
- sync/async communication between functions
- functions can be triggered from a legacy system
- react on custom events
- dynamically configure backend functions

### Data pipeline systems

- dynamically configure backend functions
- functions can be triggered from a legacy system
- react on custom events
- pull events from different systems (Kafka, RabbitMQ) via connectors

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
