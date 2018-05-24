# Reliability Guarantees

## Events are not durable

The event received by Event Gateway is stored only in memory, it's not persisted to disk before processing. This means that in case of hardware failure or software crash the event may not be delivered to the subscriber. For a synchronous subscription (`http`) it can manifest as error message returned to the requester. For asynchronous custom event with multiple subscribers it means that the event may not be delivered to all of the subscribers.

## Events are delivered _at most once_

Event Gateway attempts delivery fulfillment for an event only once and consequently any event received successfully by the Event Gateway is guaranteed to be received by the subscriber _at most once_. That said, the nature of Event Gateway provider implementation could result in retries under specific circumstances, but these should not cause delivering the same event multiple times. For example, Providers for AWS Services that use the AWS SDK are subject to auto retry logic that's built into the SDK ([AWS documentation on API retries](https://docs.aws.amazon.com/general/latest/gr/api-retries.html)).

AWS Lambda provider uses `RequestResponse` invocation type which means that retry logic for asynchronous AWS events doesn't apply here. Among others it means, that failed deliveries of custom events are not sent to DLQ. Please find more information in [Understanding Retry Behavior](https://docs.aws.amazon.com/lambda/latest/dg/retries-on-errors.html), "Synchronous invocation" section.