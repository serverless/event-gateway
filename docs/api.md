# API documentation

This document contains the API documentation for the Events API and the Configuration API in the Event Gateway. You can also find links to OpenAPI specs for these APIs.

## Contents

1.  [Events API](#events-api)
1.  [Configuration API](#configuration-api)

## Events API

[OpenAPI spec](./openapi/openapi-events-api.yaml)

The Event Gateway exposes an API for emitting events. Events API can be used for emitting custom event, HTTP events and
for invoking function. By default Events API runs on `:4000` port.

### Event Definition

All data that passes through the Event Gateway is formatted as a CloudEvent, based on CloudEvent v0.1 schema:

Example:

```json
{
  "event-type": "myapp.user.created",
  "event-id": "66dfc31d-6844-42fd-b1a7-a489a49f65f3",
  "cloud-events-version": "0.1",
  "source": "", // TBD
  "event-time": "1990-12-31T23:59:60Z",
  "data": { "foo": "bar" },
  "content-type": "application/json"
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

HTTP subscription response depends on [response object](#respond-to-an-http-event) returned by the backing function. In case of failure during function invocation following error response are possible:

* `404 Not Found` if there is no backing function registered for requested HTTP endpoint
* `500 Internal Server Error` if the function invocation failed or the backing function didn't return [HTTP response object](#respond-to-an-http-event)

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

* `200 OK` with payload returned by invoked function
* `404 Not Found` if there is no function registered or `invoke` subscription created for requested function
* `500 Internal Server Error` if the function invocation failed

### CORS

Events API supports CORS requests which means that any origin can emit a custom event. In case of `http` events CORS is
configured per-subscription basis.

## Configuration API

[OpenAPI spec](./openapi/openapi-config-api.yaml)

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

#### Update Subscription

**Endpoint**

`PUT <Configuration API URL>/v1/spaces/<space>/subscriptions/<subscription ID>`

**Request**

_Note that `event`, `functionId`, `path`, and `method` may not be updated in an UpdateSubscription call._

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

* `200 Created` on success
* `400 Bad Request` on validation error
* `404 Not Found` if subscription doesn't exist

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
