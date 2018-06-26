# API

The Event Gateway has two APIs: the Configuration API for registering functions and subscriptions, and the runtime Events API for sending events into the Event Gateway.
This document contains the API documentation for both Events and Configuration APIs. You can also find links to OpenAPI specs for these APIs.

## Contents

1.  [Events API](#events-api)
    1. [Event Definition](#event-definition)
    1. [How To Emit an Event](#how-to-emit-an-event)
    1. [HTTP Request Event](#http-request-event)
    1. [CORS](#cors)
    1. [Legacy Mode](#legacy-mode)
1.  [Configuration API](#configuration-api)
    1. [Event Types](#event-types)
        1. [Crate Event Type](#create-event-type)
        1. [Update Event Type](#update-event-type)
        1. [Delete Event Type](#delete-event-type)
        1. [List Event Types](#list-event-types)
        1. [Get Event Type](#get-event-type)
    1. [Functions](#functions)
        1. [Register Function](#register-function)
        1. [Update Function](#update-function)
        1. [Delete Function](#delete-function)
        1. [List Functions](#list-functions)
        1. [Get Function](#get-function)
    1. [Subscriptions](#subscriptions)
        1. [Create Subscription](#create-subscription)
        1. [Update Subscription](#update-subscription)
        1. [Delete Subscription](#delete-subscription)
        1. [List Subscriptions](#list-subscriptions)
        1. [Get Subscription](#get-subscription)
    1. [CORS](#cors-1)
        1. [Create CORS Configuration](#create-cors-configuration)
        1. [Update CORS Configuration](#update-cors-configuration)
        1. [Delete CORS Configuration](#delete-cors-configuration)
        1. [List CORS Configurations](#list-cors-configurations)
        1. [Get CORS Configuration](#get-cors-configuration)
    1. [Prometheus Metrics](#prometheus-metrics)
    1. [Status](#status)

## Events API

The Event Gateway exposes an API for emitting events. By default Events API runs on `:4000` port.

### Event Definition

All data that passes through the Event Gateway is formatted as a CloudEvent, based on [CloudEvents v0.1 schema](https://github.com/cloudevents/spec/blob/master/spec.md).

Example:

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

### How To Emit an Event

Creating a subscription requires `path` (default: `/`), `method` (default: `POST`) and `eventType`. `path` indicates path under which you can send the event.

**Endpoint**

`POST <Events API URL>/<Subscription Path>`

**Request**

CloudEvents payload

**Response**

Status code:

* `202 Accepted` - this status code is returned if there is no [`sync` subscription](./subscription-types.md) defined. Otherwise, status code is controlled by function synchronously subscribed on this endpoint.

### HTTP Request Event

Not all data are events that's why Event Gateway has a special, built-in `http.request` event type that enables subscribing to
raw HTTP requests. It's especially helpful for building REST APIs or supporting legacy payloads. `http.request` event is a
CloudEvent created by Event Gateway where `data` field has the following structure:

* `path` - `string` - request path
* `method` - `string` - request method
* `headers` - `object` - request headers
* `host` - `string` - request host
* `query` - `object` - query parameters
* `params` - `object` - matched path parameters
* `body` - depends on `Content-Type` header - request payload

### CORS

By default cross-origin resource sharing (CORS) is disabled. CORS is configured per-method/path basis using
[CORS Configuration API](#cors-1).

Event Gateway handles preflight `OPTIONS` requests for you. You don't need to setup subscription for `OPTIONS` method
because the Event Gateway will respond with all appropriate headers.

### Legacy Mode

*Legacy mode is deprecated and will be removed in upcoming releases.*

In legacy mode, Event Gateway is able to recognize event type based on `Event` header. If the event is not formatted according to CloudEvents specification Event Gateway looks for this header and creates CloudEvent internally. In this case, whole request body is put into `data` field.

#### Event Data Type

The MIME type of the data block can be specified using the `Content-Type` header (by default it's
`application/octet-stream`). This allows the Event Gateway to understand how to deserialize the data block if it needs
to. In case of `application/json` type the Event Gateway passes JSON payload to the target functions. In any other case
the data block is base64 encoded.

## Configuration API

[OpenAPI spec](./openapi/openapi-config-api.yaml)

The Event Gateway exposes a RESTful JSON configuration API. By default Configuration API runs on `:4001` port.

### Event Types

#### Create Event Type

**Endpoint**

`POST <Configuration API URL>/v1/spaces/<space>/eventtypes`

**Request**

JSON object:

* `name` - `string` - required, event type name
* `authorizerId` - `string` - authorizer function ID
* `metadata` - `object` - arbitrary metadata

**Response**

Status code:

* `201 Created` on success
* `400 Bad Request` on validation error

JSON object:

* `space` - `string` - space name
* `name` - `string` - event type name
* `authorizerId` - `string` - authorizer function ID
* `metadata` - `object` - arbitrary metadata

---

#### Update Event Type

**Endpoint**

`PUT <Configuration API URL>/v1/spaces/<space>/eventtypes/<event type name>`

**Request**

JSON object:

* `authorizerId` - `string` - authorizer function ID
* `metadata` - `object` - arbitrary metadata

**Response**

Status code:

* `200 OK` on success
* `400 Bad Request` on validation error or if the authorizer function doesn't exist
* `404 Not Found` if event type doesn't exist

JSON object:

* `space` - `string` - space name
* `name` - `string` - event type name
* `authorizerId` - `string` - authorizer function ID
* `metadata` - `object` - arbitrary metadata

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

#### List Event Types

**Endpoint**

`GET <Configuration API URL>/v1/spaces/<space>/eventtypes`

**Query Parameters**

Endpoint allows filtering list of returned object with filters passed as query parameters. Currently, filters can only use metadata properties e.g. `metadata.service=usersService`.

**Response**

Status code:

* `200 OK` on success

JSON object:

* `eventTypes` - `array` of `object` - event types:
  * `space` - `string` - space name
  * `name` - `string` - event type name
  * `authorizerId` - `string` - authorizer function ID
  * `metadata` - `object` - arbitrary metadata

---

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
* `metadata` - `object` - arbitrary metadata


### Functions

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
* `metadata` - `object` - arbitrary metadata

**Response**

Status code:

* `201 Created` on success
* `400 Bad Request` on validation error

JSON object:

* `space` - `string` - space name
* `functionId` - `string` - function ID
* `provider` - `object` - provider specific information about a function
* `metadata` - `object` - arbitrary metadata

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
* `metadata` - `object` - arbitrary metadata

**Response**

Status code:

* `200 OK` on success
* `400 Bad Request` on validation error
* `404 Not Found` if function doesn't exist

JSON object:

* `space` - `string` - space name
* `functionId` - `string` - function ID
* `provider` - `object` - provider specific information about a function
* `metadata` - `object` - arbitrary metadata

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

#### List Functions

**Endpoint**

`GET <Configuration API URL>/v1/spaces/<space>/functions`

**Query Parameters**

Endpoint allows filtering list of returned object with filters passed as query parameters. Currently, filters can only use metadata properties e.g. `metadata.service=usersService`.

**Response**

Status code:

* `200 OK` on success

JSON object:

* `functions` - `array` of `object` - functions:
  * `space` - `string` - space name
  * `functionId` - `string` - function ID
  * `provider` - `object` - provider specific information about a function
  * `metadata` - `object` - arbitrary metadata

---

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
* `metadata` - `object` - arbitrary metadata

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
* `metadata` - `object` - arbitrary metadata

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
* `metadata` - `object` - arbitrary metadata

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
* `metadata` - `object` - arbitrary metadata

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
* `metadata` - `object` - arbitrary metadata

---

#### Delete Subscription

**Endpoint**

`DELETE <Configuration API URL>/v1/spaces/<space>/subscriptions/<subscription ID>`

**Response**

Status code:

* `204 No Content` on success
* `404 Not Found` if subscription doesn't exist

---

#### List Subscriptions

**Endpoint**

`GET <Configuration API URL>/v1/spaces/<space>/subscriptions`

**Query Parameters**

Endpoint allows filtering list of returned object with filters passed as query parameters. Currently, filters can only use metadata properties e.g. `metadata.service=usersService`.

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
  * `metadata` - `object` - arbitrary metadata

---

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
* `metadata` - `object` - arbitrary metadata

### CORS

#### Create CORS Configuration

**Endpoint**

`POST <Configuration API URL>/v1/spaces/<space>/cors`

**Request**

* `method` - `string` - endpoint method
* `path` - `string` - endpoint path
* `allowedOrigins` - `array` of `string` - list of allowed origins. An origin may contain a wildcard (\*) to replace 0 or more characters (i.e.: http://\*.domain.com), default: `*`
* `allowedMethods` - `array` of `string` - list of allowed methods, default: `HEAD`, `GET`, `POST`
* `allowedHeaders` - `array` of `string` - list of allowed headers, default: `Origin`, `Accept`, `Content-Type`
* `allowCredentials` - `bool` - allow credentials, default: false
* `metadata` - `object` - arbitrary metadata

**Response**

Status code:

* `201 Created` on success
* `400 Bad Request` on validation error

JSON object:

* `space` - `string` - space name
* `corsId` - `string` - CORS configuration ID
* `method` - `string` - endpoint method
* `path` - `string` - endpoint path
* `allowedOrigins` - `array` of `string` - list of allowed origins
* `allowedMethods` - `array` of `string` - list of allowed methods
* `allowedHeaders` - `array` of `string` - list of allowed headers
* `allowCredentials` - `boolean` - allow credentials
* `metadata` - `object` - arbitrary metadata

---

#### Update CORS Configuration

**Endpoint**

`PUT <Configuration API URL>/v1/spaces/<space>/cors/<CORS ID>`

**Request**

_Note that `method`, and `path` may not be updated in an UpdateCORS call._

* `method` - `string` - endpoint method
* `path` - `string` - endpoint path
* `allowedOrigins` - `array` of `string` - list of allowed origins
* `allowedMethods` - `array` of `string` - list of allowed methods
* `allowedHeaders` - `array` of `string` - list of allowed headers
* `allowCredentials` - `boolean` - allow credentials
* `metadata` - `object` - arbitrary metadata

**Response**

Status code:

* `200 Created` on success
* `400 Bad Request` on validation error
* `404 Not Found` if CORS configuration doesn't exist

JSON object:

* `space` - `string` - space name
* `corsId` - `string` - CORS configuration ID
* `method` - `string` - endpoint method
* `path` - `string` - endpoint path
* `allowedOrigins` - `array` of `string` - allowed origins
* `allowedMethods` - `array` of `string` - allowed methods
* `allowedHeaders` - `array` of `string` - allowed headers
* `allowCredentials` - `boolean` - allow credentials
* `metadata` - `object` - arbitrary metadata

---

#### Delete CORS Configuration

**Endpoint**

`DELETE <Configuration API URL>/v1/spaces/<space>/cors/<CORS ID>`

**Response**

Status code:

* `204 No Content` on success
* `404 Not Found` if CORS configuration doesn't exist

---

#### List CORS Configurations

**Endpoint**

`GET <Configuration API URL>/v1/spaces/<space>/cors`

**Query Parameters**

Endpoint allows filtering list of returned object with filters passed as query parameters. Currently, filters can only use metadata properties e.g. `metadata.service=usersService`.

**Response**

Status code:

* `200 OK` on success

JSON object:

* `cors` - `array` of `object` - CORS configurations
  * `space` - `string` - space name
  * `corsId` - `string` - CORS configuration ID
  * `method` - `string` - endpoint method
  * `path` - `string` - endpoint path
  * `allowedOrigins` - `array` of `string` - allowed origins
  * `allowedMethods` - `array` of `string` - allowed methods
  * `allowedHeaders` - `array` of `string` - allowed headers
  * `allowCredentials` - `boolean` - allow credentials
  * `metadata` - `object` - arbitrary metadata

---

#### Get CORS Configuration

**Endpoint**

`GET <Configuration API URL>/v1/spaces/<space>/cors/<CORS ID>`

**Response**

Status code:

* `200 OK` on success
* `404 NotFound` if CORS configuration doesn't exist

JSON object:

* `space` - `string` - space name
* `corsId` - `string` - CORS configuration ID
* `method` - `string` - endpoint method
* `path` - `string` - endpoint path
* `allowedOrigins` - `array` of `string` - allowed origins
* `allowedMethods` - `array` of `string` - allowed methods
* `allowedHeaders` - `array` of `string` - allowed headers
* `allowCredentials` - `boolean` - allow credentials
* `metadata` - `object` - arbitrary metadata

### Prometheus Metrics

Endpoint exposing [Prometheus metrics](./prometheus-metrics.md).

**Endpoint**

`GET <Configuration API URL>/metrics`

### Status

Dummy endpoint (always returning `200 OK` status code) for checking if the event gateway instance is running.

**Endpoint**

`GET <Configuration API URL>/v1/status`
