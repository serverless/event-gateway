# Use-Cases/Scenarios

## HTTP Endpoint/REST API

Framework Example

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

SDK Example

```javascript
// Step 1. Register A Function (done only once per function)
sdk.registerFunction("creataUser", {
  awsLambda: {
    arn: "xxx",
    region: "us-west-2",
    version: 2,
    accessKeyId: "xxx",
    secretAccessKey: "xxx"
  }
})

// Step 2. Subscribe A Function to "http" Event

sdk.subscribe({
  function: "createUser",
  event: "http",
  method: "GET",
  path: "users"
})

// "http" is a built-in sync event that also allows to configure "method" and "path"
```

## Publishing & Subscribing to Custom Events

Framework Example

```yaml
functions:
  sendWelcomeEmail:
    handler: emails.welcome
    events:
      - userCreated
```

SDK Example

```javascript
sdk.subscribe({
  function: "sendWelcomeEmail",
  event: "userCreated"
})
```

## Sharing Events Between Services (Inside Org/Team/App)

Communication inside the same Org/Team/App is open. Services can subscribe to events emitted by other services.

```javascript
// users-service emits "userCreated" event

sdk.emit({
  event: "userCreated",
  data: {
    id: "1",
    name: "Foo"
  }
})

// emails-service can subscribe to userCreated event via the framework

functions:
  sendWelcomeEmail:
    handler: emails.welcome
    events:
      - userCreated

// Or via SDK

sdk.subscribe({
  function: "emails-service/sendWelcomeEmail",
  event: "userCreated"
})
```

## Sharing Events With A Team (Inside Org)

Same as above.

## Sharing Events With An End User (Outside Org, Replaces Webhooks)

Cocacola company wants to share "fooBar" event with Nike. They are running on SaaS hosted Event Gateways.

Coca cola admin needs to grant access for creating subscription on "fooBar" event for Nike user.

```javascript
var sdk = new SDK("cocacola.serverless.com", "myapikey")

sdk.grant({
  gateway: "nike.serverless.com",
  action: "create-subscription",
  event: "fooBar"
})
```

Nike can subscribe to that event on CocaCola gateway

```yaml
gateways:
	cocacola: cocacola.serverless.com

functions:
  bar:
    handler: bar.bar
    events:
      - cocacola.fooBar
```

## Allow Another Team To Publish Events

If both teams are deployed on the same gateway there is no need to explicitly grant permissions.

## Allow An End User To Publish Events

Cocacola admin allows Nike's gateway to publish "fooBar" event into Cocacola's gateway.

```javascript
var sdk = new SDK("cocacola.serverless.com", "myapikey")

sdk.grant({
  gateway: "nike.serverless.com",
  action: "publish-subscription",
  event: "fooBar"
})
```

## Log/process HTTP Requests coming in on an Endpoint

User is able to subscribe to system "gateway.event.emitted" event

```javascript
sdk.subscribe({
  function: "logRequests"
  event: "gateway.event.emitted",
  filters: {
    event: "http",
  	path: "/users"
  }
})
```

## Log/process HTTP Requests coming in on any Endpoint subpath



## Subscribe to Events across the entire Event Gateway

```javascript
sdk.subscribe({
  function: "sendWelcomeEmail",
  event: "*"
})
```

## Subscribe to Event Gateway System Events

The Event Gateway exposes internal events:

- gateway.event.emitted
- gateway.function.registered
- gateway.function.unregistered
- gateway.function.invoked
- gateway.subscription.created
- gateway.subscription.deleted
- ...

## Canary Deployments

User is able to register a group function that uses already registered function as a backing functions

SDK Example

```javascript
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
})
```

## Retries

Retries are defined during a function registration.

SDK Example

```javascript
sdk.registerFunction("creataUser", {
  awsLambda: {
    arn: "xxx",
    region: "us-west-2",
    version: 2,
    accessKeyId: "xxx",
    secretAccessKey: "xxx"
  },
  retries: {
    retry: 3,
    exponentialBackoff: true
  }
})
```

## Chaining synchronous function calls

User is able to register a chain function that uses already registered function as a backing functions

```javascript
sdk.registerFunction("hello-world-chain", {
  chain: {
    functions: ["hello-world-step1", "hello-world-step2"]
  }
})
```

