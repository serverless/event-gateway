# Subscription Types

Event Gateway supports two subscription types: `async` and `sync`.

## `async` subscription

`async` subscription implements lightweight pub/sub system. There can be many async subscriptions listening to the same event type, on the same path and HTTP method. The function is asynchronously invoked by the Event Gateway.

## `sync` subscription

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

Key and value of matched parameters are passed to a function in an [HTTP Request Event](./api.md#http-request-event) under `params` field.

**Wildcard parameters**

Special type of path parameter is wildcard parameter. It's a path segment prefixed with `*`. Wildcard parameter can only
be specified at the end of the path and will match every character till the end of the path. For examples
parameter `/users/*userpath` for request path `/users/group1/user1` will match `group1/user1` as a `userpath` parameter.