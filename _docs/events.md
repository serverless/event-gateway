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

### Pub-Sub
Topics can be configured to receive the input or output of a function.
A function may feed multiple topics.
A topic may be fed by multiple functions.
Multiple functions may subscribe to a topic.

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

### Synchronous List
A chain of functions that are executed one after the other by the gateway,
the input to the gateway used as the input for the first function,
the output of the first used as input to the second, and so on,
with the output of the final function being returned to the caller.
The same function may appear several times in the list, and the
gateway will not trigger a cycle.

### Synchronous Tree
Similar to synchronous list, except the results of a function may be
sent to multiple functions, branching out asynchronously.
A single function is specified as the output function, which will
ultimately populate the final response to the caller.

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
