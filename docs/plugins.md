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
