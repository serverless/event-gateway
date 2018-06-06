# API

The Event Gateway has two APIs: the Configuration API for registering functions and subscriptions, and the runtime Events API for sending events into the Event Gateway.
This document contains the API documentation for both Events and Configuration APIs. You can also find links to OpenAPI specs for these APIs.

## Contents

1.  [Events API](#events-api)
    1. [Event Registry](#event-registry)
    1. [Event Definition - CloudEvents](#event-definition---cloudevents)
        1. [CloudEvents Example](#cloudevents-example)
    1. [Subscription Types](#subscription-types)
        1. [`async` subscription](#async-subscription)
        2. [`sync` subscription](#sync-subscription)
    1. [Authorization](#authorization)
        1. [Invocation Payload](#invocation-payload)
        1. [Invocation Result](#invocation-result)
    1. [HTTP Request Event](#http-request-event)
    1. [How To Emit an Event](#how-to-emit-an-event)
    1. [Legacy Mode](#legacy-mode)
1.  [Configuration API](#configuration-api)
    1. [Event Registry](#event-registry)
        1. [Register Event Type](#register-event-type)
        1. [Delete Event Type](#delete-event-type)
        1. [Get Event Types](#get-event-types)
        1. [Get Event Type](#get-event-type)
    1. [Function Discovery](#function-discovery)
        1. [Register Function](#register-function)
        1. [Update Function](#update-function)
        1. [Delete Function](#delete-function)
        1. [Get Functions](#get-functions)
        1. [Get Function](#get-function)
    1. [Subscriptions](#subscriptions)
        1. [Create Subscription](#create-subscription)
        1. [Update Subscription](#update-subscription)
        1. [Delete Subscription](#delete-subscription)
        1. [Get Subscriptions](#get-subscriptions)
        1. [Get Subscription](#get-subscription)
    1. [Prometheus Metrics](#prometheus-metrics)
    1. [Status](#status)

## Events API

[OpenAPI spec](./openapi/openapi-events-api.yaml)

The Event Gateway exposes an API for emitting events. By default Events API runs on `:4000` port.

### Event Registry

Event Registry is a single source of truth about events occuring in the space. Every event emitted to a space has to have type registered beforehand.

### Event Definition - CloudEvents

Event Gateway has first-class support for [CloudEvents](https://cloudevents.io/). It means few things.

First of all, if the event emitted to the Event Gateway is in CloudEvents format, the Event Gateway is able to recognize it and trigger proper subscriptions based on event type specified in the event. Event Gateway supports both Structured Content and Binary Content modes described in [HTTP Transport Binding spec](https://github.com/cloudevents/spec/blob/master/http-transport-binding.md).

Secondly, there is a special, built-in [HTTP Request Event](#http-request-event) type allowing reacting to raw HTTP requests that are not formatted according to CloudEvents spec. This event type can be especially helpful for building REST APIs.

Currently, Event Gateway supports [CloudEvents v0.1 schema](https://github.com/cloudevents/spec/blob/master/spec.md) specification.

#### CloudEvents Example

```json
{
  "eventType": "myapp.user.created",
  "eventID": "66dfc31d-6844-42fd-b1a7-a489a49f65f3",
  "cloudEventsVersion": "0.1",
  "source": "https://serverless.com/event-gateway/#transformationVersion=0.1",
  "eventTime": "1990-12-31T23:59:60Z",
  "data": { "foo": "bar" },
  "contentType": "application/json"
}
```

### Subscription Types

Event Gateway support two subscription types: `async` and `sync`.

#### `async` subscription

`async` subscription implements lightweight pub/sub system. There can be many async subscriptions listening to the same event type, on the same path and HTTP method. The function is asynchronously invoked by the Event Gateway.

#### `sync` subscription

In case of `sync` subscription invoked function can control the HTTP response returned by the Event Gateway. Because of that, there can be only one `sync` subscription for a path, HTTP method, and event type tuple.

Status code, headers and response body can be controlled by returning JSON object with the following fields:

* `statusCode` - `int` - response status code, default: 200
* `headers` - `object` - response headers
* `body` - `string` - response body. Currently, the Event Gateway supports only string responses.

If the function invocation failed or the backing function didn't return JSON object in the above format Event Gateway returns `500 Internal Server Error`.

**Path parameters**

The Event Gateway allows creating `sync` subscription with parameterized paths. Every path segment prefixed with `:` is
treated as a parameter, e.g. `/users/:id`.

The Event Gateway prevents from creating subscriptions in following conflicting situations:

* registering static path when there is parameterized path registered already (`/users/:id` vs. `/users/foo`)
* registering parameterized path with different parameter name (`/users/:id` vs. `/users/:name`)

Key and value of matched parameters are passed to a function in an [HTTP Request Event](#http-request-event) under `params` field.

**Wildcard parameters**

Special type of path parameter is wildcard parameter. It's a path segment prefixed with `*`. Wildcard parameter can only
be specified at the end of the path and will match every character till the end of the path. For examples
parameter `/users/*userpath` for request path `/users/group1/user1` will match `group1/user1` as a `userpath` parameter.

### Authorization

Event Type can define authorizer function that will be called before calling a subscribed function. Authorizer function is a function registered in Function Discovery beforehand.

#### Invocation Payload

The authorizer function is invoked with a special payload. The function has access to the whole request object because different parts of the request can be required for running authorization logic (e.g. API key can be stored in different headers or query params). The invocation payload is a JSON object with the following structure:

- `event` - `object` - event received and parsed by EG
- `request` - `object` - original HTTP request to the EG, this field is exactly the same as HTTP event, including body, which in case of CloudEvent will be exactly the same as event field

#### Invocation Result

The authorize function is expected to return authorization response JSON object with the following structure:

- `authorization` - `object` - object containing authorization data, required if authorization is successful. Fields:
  - `principalId` - `string` - required if authorization is successful, the principal user identification associated with the token sent in the request
  - `context` - `map[string]string` - arbitrary data that will be accessible by the downstream function
- `error` - `object` - error information, required if authorization is unsuccessful. Fields:
  - `message` - `string` - authorization error message

If the authorization is successful `error` field has to be null or not defined otherwise Event Gateway treats authorization process as unsuccessful and `403 Forbidden` error is returned to the client assuming that there was sync subscription defined. `authorization` object is attached to CloudEvent Extensions field under `eventgateway.authorization` key.

If Event Gateway will not be able to parse the response from authorizer function or there will be error during invocation authorization process is considered as unsuccessful.

### HTTP Request Event

`http.request` event is a built-in event type that wraps raw HTTP requests. Not all data are events that's why this type of event is especially helpful for building REST APIs or supporting legacy payloads. `http.request` event is a CloudEvent created by Event Gateway where `data` field has the following structure:

* `path` - `string` - request path
* `method` - `string` - request method
* `headers` - `object` - request headers
* `host` - `string` - request host
* `query` - `object` - query parameters
* `params` - `object` - matched path parameters
* `body` - depends on `Content-Type` header - request payload

### How To Emit an Event

Creating a subscription requires `path` (default: `/`), `method` (default: `POST`) and `eventType`. `path` indicates path under which you can send the event.

**Endpoint**

`POST <Events API URL>/<Subscription Path>`

**Request**

CloudEvents payload

**Response**

Status code:

* `202 Accepted` - if there is no `sync` subscription. Otherwise, status code is controlled by function synchronously subscribed on this endpoint.

### Legacy Mode

In legacy mode, Event Gateway is able to recognize event type based on `Event` header. If the event is not formatted according to CloudEvents specification Event Gateway looks for this header and creates CloudEvent internally. In this case, whole request body is put into `data` field.

#### Event Data Type

The MIME type of the data block can be specified using the `Content-Type` header (by default it's
`application/octet-stream`). This allows the Event Gateway to understand how to deserialize the data block if it needs
to. In case of `application/json` type the Event Gateway passes JSON payload to the target functions. In any other case
the data block is base64 encoded.

## Configuration API

[OpenAPI spec](./openapi/openapi-config-api.yaml)

The Event Gateway exposes a RESTful JSON configuration API. By default Configuration API runs on `:4001` port.

### Event Registry

#### Register Event Type

**Endpoint**

`POST <Configuration API URL>/v1/spaces/<space>/eventtypes`

**Request**

JSON object:

* `name` - `string` - required, event type name
* `authorizerId` - `string` - authorizer function ID

**Response**

Status code:

* `201 Created` on success
* `400 Bad Request` on validation error

JSON object:

* `space` - `string` - space name
* `name` - `string` - event type name
* `authorizerId` - `string` - authorizer function ID

---

#### Delete Event Type

Delete event type. This operation fails if there is at least one subscription using the event type.

**Endpoint**

`DELETE <Configuration API URL>/v1/spaces/<space>/eventtypes/<event type name>`

**Response**

Status code:

* `204 No Content` on success
* `400 Bad Request` if there are subscriptions using the event type
* `404 Not Found` if event type doesn't exist

---

#### Get Event Types

**Endpoint**

`GET <Configuration API URL>/v1/spaces/<space>/eventtypes`

**Response**

Status code:

* `200 OK` on success

JSON object:

* `eventTypes` - `array` of `object` - event types:
  * `space` - `string` - space name
  * `name` - `string` - event type name
  * `authorizerId` - `string` - authorizer function ID

#### Get Event Type

**Endpoint**

`GET <Configuration API URL>/v1/spaces/<space>/eventtypes/<event type name>`

**Response**

Status code:

* `200 OK` on success
* `404 Not Found` if event type doesn't exist

JSON object:

* `space` - `string` - space name
* `name` - `string` - event type name
* `authorizerId` - `string` - authorizer function ID


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
  * for AWS SQS connector:
    * `queueUrl` - `string` - required, AWS SQS Queue URL
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

* `type` - `string` - subscription type, `sync` or `async`
* `eventType` - `string` - event type
* `functionId` - `string` - ID of function to receive events
* `path` - `string` - optional, URL path under which events (HTTP requests) are accepted, default: `/`
* `method` - `string` - optional, HTTP method that accepts requests, default: `POST`
* `cors` - `object` - optional, by default CORS is disabled for `sync` subscriptions. When set to empty object CORS configuration will use default values for all fields below. Available fields:
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
* `type` - `string` - subscription type
* `eventType` - `string` - event type
* `functionId` - function ID
* `method` - `string` - HTTP method that accepts requests
* `path` - `string` - path that accepts requests, starts with `/`
* `cors` - `object` - optional, CORS configuration

---

#### Update Subscription

**Endpoint**

`PUT <Configuration API URL>/v1/spaces/<space>/subscriptions/<subscription ID>`

**Request**

_Note that `type`, `eventType`, `functionId`, `path`, and `method` may not be updated in an UpdateSubscription call._

* `type` - `string` - subscription type, `sync` or `async`
* `eventType` - `string` - event type
* `functionId` - `string` - ID of function to receive events
* `path` - `string` - optional, URL path under which events (HTTP requests) are accepted, default: `/`
* `method` - `string` - optional, HTTP method that accepts requests, default: `POST`
* `cors` - `object` - optional, by default CORS is disabled for `sync` subscriptions. When set to empty object CORS configuration will use default values for all fields below. Available fields:
  * `origins` - `array` of `string` - list of allowed origins. An origin may contain a wildcard (\*) to replace 0 or more characters (i.e.: http://\*.domain.com), default: `*`
  * `methods` - `array` of `string` - list of allowed methods, default: `HEAD`, `GET`, `POST`
  * `headers` - `array` of `string` - list of allowed headers, default: `Origin`, `Accept`, `Content-Type`
  * `allowCredentials` - `bool` - default: false

**Response**

Status code:

* `200 Created` on success
* `400 Bad Request` on validation error
* `404 Not Found` if subscription doesn't exist

JSON object:

* `space` - `string` - space name
* `subscriptionId` - `string` - subscription ID
* `type` - `string` - subscription type
* `eventType` - `string` - event type
* `functionId` - function ID
* `method` - `string` - HTTP method that accepts requests
* `path` - `string` - path that accepts requests, starts with `/`
* `cors` - `object` - optional, CORS configuration

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
  * `type` - `string` - subscription type
  * `eventType` - `string` - event type
  * `functionId` - function ID
  * `method` - `string` - HTTP method that accepts requests
  * `path` - `string` - path that accepts requests, starts with `/`
  * `cors` - `object` - optional, CORS configuration

#### Get Subscription

**Endpoint**

`GET <Configuration API URL>/v1/spaces/<space>/subscriptions/<subscription ID>`

**Response**

Status code:

* `200 OK` on success
* `404 NotFound` if subscription doesn't exist

JSON object:

* `space` - `string` - space name
* `subscriptionId` - `string` - subscription ID
* `type` - `string` - subscription type
* `eventType` - `string` - event type
* `functionId` - function ID
* `method` - `string` - HTTP method that accepts requests
* `path` - `string` - path that accepts requests, starts with `/`
* `cors` - `object` - optional, CORS configuration

### Prometheus Metrics

Endpoint exposing [Prometheus metrics](./prometheus-metrics.md).

**Endpoint**

`GET <Configuration API URL>/metrics`

### Status

Dummy endpoint (always returning `200 OK` status code) for checking if the event gateway instance is running.

**Endpoint**

`GET <Configuration API URL>/v1/status`
