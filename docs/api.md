# API

The Event Gateway has two APIs: the Configuration API for registering functions and subscriptions, and the runtime Events API for sending events into the Event Gateway.
This document contains the API documentation for both Events and Configuration APIs. You can also find links to OpenAPI specs for these APIs.

## Contents

1.  [Events API](#events-api)
    1. [Event Definition](#event-definition)
    1. [How To Emit an Event](#how-to-emit-an-event)
    1. [HTTP Request Event](#http-request-event)
    1. [Legacy Mode](#legacy-mode)
1.  [Configuration API](#configuration-api)
    1. [Event Types](#event-types)
        1. [Register Event Type](#register-event-type)
        1. [Update Event Type](#update-event-type)
        1. [Delete Event Type](#delete-event-type)
        1. [Get Event Types](#get-event-types)
        1. [Get Event Type](#get-event-type)
    1. [Functions](#functions)
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

#### Update Event Type

**Endpoint**

`PUT <Configuration API URL>/v1/spaces/<space>/eventtypes/<event type name>`

**Request**

JSON object:

* `authorizerId` - `string` - authorizer function ID

**Response**

Status code:

* `200 OK` on success
* `400 Bad Request` on validation error or if the authorizer function doesn't exist
* `404 Not Found` if event type doesn't exist

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

### Prometheus Metrics

Endpoint exposing [Prometheus metrics](./prometheus-metrics.md).

**Endpoint**

`GET <Configuration API URL>/metrics`

### Status

Dummy endpoint (always returning `200 OK` status code) for checking if the event gateway instance is running.

**Endpoint**

`GET <Configuration API URL>/v1/status`
