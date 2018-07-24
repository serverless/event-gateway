# System Events and Plugin System

## System Events

System Events are special type of events emitted by the Event Gateway instance internally. They are emitted on each stage
of event processing flow starting from receiving event to function invocation end. Those events are:

* `eventgateway.event.received` - the event is emitted when an event was received by Events API. Data fields:
  * `event` - event payload
  * `path` - Events API path
  * `headers` - HTTP request headers
* `eventgateway.function.invoking` - the event emitted before invoking a function. Data fields:
  * `space` - space name
  * `event` - event payload
  * `functionId` - registered function ID
* `eventgateway.function.invoked` - the event emitted after successful function invocation. Data fields:
  * `space` - space name
  * `event` - event payload
  * `functionId` - registered function ID
  * `result` - function response
* `eventgateway.function.invocationFailed` - the event emitted after failed function invocation. Data fields:
  * `space` - space name
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

For more details, see [the example plugin](../plugin/example).
