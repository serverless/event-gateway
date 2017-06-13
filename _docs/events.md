# Event Specification

## Warning: On Cycles
A cycle means the entire system grinds to a halt, as the gateway infinitely passes
events back and forth, getting hotter and hotter as more events enter the cycle.
We may want to consider adding a defensive TTL field, similar to that
used in IP packets, which is decremented at each subsequent call
of the gateway, being dropped when it hits 0.
TTL's do not prevent cycles, but they allow our system to stay alive when they occur.
They will occur by accident, and we should be paying attention for them, because
they will destroy our product if we can't handle them when they do occur.

## Event Patterns

### Call-Response
A function may be called by name through the gateway.
A function may call other functions by name through the gateway.

Code must be changed to invoke functions in different patterns.

examples:

* A function for handling an HTTP GET request for a social media
homepage issues requests through the gateway to call different
functions for retrieving a list of unread messages, a list
of posts by followed users, and a set of advertisements to display
* If we want to add an authentication check, we deploy a new 
function that calls an authentication function through the gateway.

### Multiple Emit
A function produces metrics and log messages during the course of
execution. However, it would dramatically slow down the function
if it tried to synchronously send them to an external system before
returning the output to the caller (the gateway). The function may
return an object in a form that the gateway knows to split into
a "main response" that is returned to the caller and "named events"
that are destined for specific topics, such as `homepage_metrics` or 
`new_user_log`. This is also a way for a function to "optionally
emit" an event to a named topic, like an error or security log.

example:

* The login function either returns a security token for a user,
or an error message. It returns an object to the gateway that
contains both the user response as well as
some combination of metrics about successful logins,
unsuccessful logins, unsuccessful IP address requestors, and
unsuccessful usernames to power a rate limiting system that
fights abuse.

### Cutover
A named function has multiple backing functions, each with a "weight"
value that determines the proportional amount of traffic each should
receive. If all are the same, they all receive requests equally.
If one is 99 and another is 1, 99% of traffic will be sent to the
one with a weight of 99.

example:

* Blue-green deploys: a new version of a function has been written
and we would like to deploy it. However, we may have missed a key
bug during testing, and we don't want to completely roll it out.
Engineers start the deploy by setting the predecessor's weight to
99, adding the new function to the same named function config,
but with a weight of 1. The engineers monitor the metrics and logs
emitted by the new function and related functionality to ensure
that the overall health of the system is uncompromised. Incrementally,
engineers increase the weight of the new function, ramping up
as they build confidence, until the new function is 100, and the old
one is 0. The old function is removed from the configuration completely
after a safe period of time has elapsed. If a problem is discovered,
the old function is immediately reset to 100 and the new function to 0.

### Pub-Sub
Topics can be configured to receive the input or output of a function.
A function may feed multiple topics.
A topic may be fed by multiple functions.
Multiple functions may subscribe to a topic.

Code does not need to change to modify the overall event graph, just
configuration in the gateway.

When the gateway receives a request to call a function by name, it checks
whether the called function is an input to one or more topics.
If the function is an input, the gateway will forward its input or
output to all subscribers according to the topic configuration.

At a certain point, one gateway process will not be able to efficiently
process the entire subscription graph of an input event.
We will measure how wide a subscription graph needs to be to benefit
from distributed execution. We may not ever need to think about this
for 99% of users, but we don't know yet, and we may need to do it
early on.

A configuration request must fail if it creates a loop in the subscription graph.
Subscription graph configuration must use atomic updates in the database to
avoid cycles at all costs.

example:

* Someone wants to build a feature that shows the top-5 most popular searches,
so they create a function that performs Misra-Gries on the stream of searches,
and optionally emits a new set of top-5 when the membership of the top 5
changes to the `top-5-searches` topic, which is subscribed to by
a function that populates a cache with the top 5 for rapid serving on a
sidebar of the homepage.

### Synchronous List
A chain of functions that are executed one after the other by the gateway,
the input to the gateway used as the input for the first function,
the output of the first used as input to the second, and so on,
with the output of the final function being returned to the caller.
The same function may appear several times in the list, and the
gateway will not trigger a cycle.

Any function can return a short-circuit response, such as an error message,
which will skip the rest of the functions and return a response directly.

This completely decouples the code of one function from another.

example:

* We define a chain of functions, each enriching a result, to render a webpage,
account lookup -> find orders -> filter orders by a query -> render html.
If we want to add authentication to this chain, we just reconfigure the
list to start with a filter that either returns a short-circuit error or
allows the list of functions to continue processing until the end.
* We want to filter new requests based on spam filters, without changing
any of the other functionality. We reconfigure the List to include a
spam filter that can short-circuit a failure message if spam is detected.

### Synchronous Tree
Similar to synchronous list, except the results of a function may be
sent to multiple functions, branching out asynchronously or conditionally.
At a branch of the tree, the gateway will either send the result to
all children, or a single one based on pattern matching the event
and forwarding in a configured way.

examples:
* A user has just entered credentials on a page. We either want to trigger
an anti-abuse set of functions in parallel if they failed to log in, or generate
a set of rich content based on their account's preferences.
* We want to start feeding requests to a new function under development but
not yet ready for deployment. We want to make sure that it does not
cause any bugs in the system, so we reconfigure the Tree to asynchronously
run a different set of functions, feeding a metric system that is used
to ensure that the overall health of the system is not compromised. After
enough confidence is gained to put it in production, we feed it to the
cutover system and slowly deploy it.

### Synchronous Graph
A Graph of functions (a [DAG](https://en.wikipedia.org/wiki/Directed_acyclic_Graph)
with one input and one output) may be specified.
It is similar to a synchronous tree, except some functions may
take more than one input. We can parallelize multiple function
inputs and feed the function a list.
The list of inputs will initially be a new concept to users, but it allows
for significant performance improvements when the inputs to a
function can be retrieved in parallel.

A node in the Graph is determined by its position, not function, and a function
may appear at multiple points of the Graph, but there may be no
loops among nodes in the Graph.

example:
* We want to generate several prerequisites for a webpage in parallel.
It would take too long to sequentially get each component of the response.
We scatter an initial HTTP GET request to 5 different functions, each
responsible for fetching and processing 1/5 of the content to display, and
then they fan-in to a single function that generates the HTML to return
to the user.

## Questions
1. Do we allow events triggered as part of a Synchronous
topology to trigger Pub-Sub responses? This could be quite
powerful, or confusing and prone to billing surprises.
It is harder to do capacity planning for.
We need to perform cycle detection before allowing a
Graph or Pub-Sub subscriber to be deployed, and what if
there is a loop that is caused by the interaction?
This needs to be carefully surfaced to the user
to avoid confusion.
1. Do we allow events from Pub-Sub to trigger the
Synchronous Topologies? Probably not, because
there is nothing to be synchronously responded to.
1. Do we allow Synchronous topologies to trigger events that
will not contribute to the final synchronous response? It's hard for us to
reason about that accurately, as one event may act on an external database
that is then read by a later function, so we should probably not
try to reduce the Synchronous topologies to the direct dependencies
of the specified output to use as a response.
1. If we allow multiple/optional return events, how can we
decouple response types from configurable topics, such that
we don't need to know about specific topics when deploying a
function that may emit events of type A, B, or C? Maybe
allow configuring a mapping from a function's event type identifiers
to one or more topics.

## Recommended Initial Functionality
Here are some things that would be good to build initially, which
will allow us to be flexible later on about what patterns we
choose to support.

1. Multiple backing functions for a single named function. Supports
cutover/blue-green deploys/canary deployments.
1. TTL fields on requests to the gateway, based on a TTL passed
to the function. Even with graph analysis, it will be
hard to guarantee that no loops will occur as long as anything
can make a request through the gateway to another function
(or itself accidentally). We need to detect which things are
triggering loops and mitigate their damage as fast as possible.
1. Multiple/optional response events. Maybe pass in a token to a function,
and if the function returns JSON/PB/etc... that includes the provided token,
we know that it is an object that may contain multiple/optional events.
1. Embrace many:many thinking for everything. Avoid things
that need to be configured for a particular destination in the
code. Push as much to configuration as possible.

