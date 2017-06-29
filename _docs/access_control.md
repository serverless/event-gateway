# Access Control Spec

## Goals

* enable organizations to share resources in a powerful, familiar, easy way
* define a target based on industry best-practices
* access-control configuration fully-decoupled from code (full separation of policy and mechanism,
  in security-speak)
* a target that is amenable to incremental development, that some core functionality can be built for
  by the time of the Emit conference.
* a priority list of features that will be implemented in a best-effort sequence

## Non-Goals

* complete implementation by the time of the Emit conference. multi-user access control usually takes
  years to implement (see kubernetes: RBAC only available as beta after several years of work, nomad:
  issue open since october 2015, docker swarm: issue open since september 2015, dcos: after a
  year of active development, only the ability to add users with no access control is finished,
  the list goes on...)

## Concepts

### Identities

An identity is a mapping between an identity ID and a set of tokens. It solves key
management problems of just supporting tokens without limiting the interoperability
of the system.

Key management is generally considered one of the most painful issues in security. Pain points:

* When using tokens without any key management, you quickly get to a point where
  the system is expensive and error-prone to manage.
* If a team has to interface with a changing set of external functions / orgs / topics, they will
  have a token for each one, and they will spend a ton of energy keeping track of the tokens.
* If a team wants to revoke access for a user, they need to have kept records (maybe an excel doc or something)
  about all of the tokens they've granted to that user, and then revoke each one.
* If an org wants to lock out a particular user, they need to have a record of every token they may have acquired,
  and revoke each one
* A user does a deploy at 5:57PM then jumps in the subway to go home. It starts misbehaving, and teams running
  other systems are the first to notice. All effected teams need to coordinate to block that system from bringing
  down their systems. They better all have up-to-date excel docs.
* People sometimes get their laptops stolen. We sometimes need to keep a user's systems running, but change
  the way that they are accessed. Key rotation is impossible with a static token, and we need to re-grant all
  privileges for all systems, then update all the excel docs. Hopefully we have time, and catch our typos. We incur
  downtime as our systems cut over to a new token.

A better way:

* Associate an ID name with a set of backing tokens.
* Grant and revoke privileges based on that ID, no need to securely transmit a secret token.
* No authentication requirements, we just have a mapping between a identity and their tokens.
* The identity/team behind the ID does not need to manage 50+ changing tokens for use with specific other services,
  they just keep their one token.
* Our security team is happy because we can enforce mandatory periodic key rotation. by supporting a mapping from
  ID to several tokens, we can let teams gracefully cut over to their new tokens, and phase out old ones, without
  needing a "hard cut-over" that requires downtime.
* It's not much more than a hashmap on our end, but it dramatically simplifies the lives of our identitys, and we do not
  impose any additional constraints. They can make their token the ID if they want to do fully-manual key management.

```
{
  id: "alice",
  tokens: [
    "XXX1",
    "XX2X"
  ]
}
```

Granting and revoking rules no longer involves messy manual management of tokens. You include
your token in your requests, and you don't need to think about which token you need to use
for each backing system.

Authentication is outside the scope of this system for now, but we could prioritize it in the future.

### Ownership

* all objects are assigned an owning identity upon creation, including new identities
* ownership is hierarchical

### Namespaces

* A Namespace is a coarse-grained sandbox in which entities can interact freely.
* An identity is a member of one or more namespaces.
* Topics, Functions, and Endpoints belong to one or more namespaces.
* All actions are possible within a namespace: publishing, subscribing and calling
* All access cross-namespace is disabled by default.
* To perform cross-namespace access, use Rules (see below, not initial dev focus before Emit)

### Rules

* Probably not going to exist before Emit. Maybe never, if Namespaces are good enough.
* Rules are defined as a tuple of (Subject, Action, Object, Environment).
* Subject is an identity or function ID.
* Object may be an identity, function, endpoint, topic, or a group of them.
* Environment may relate to a logical namespace, datacenter, geographical region, the event gateway's
  conception of time, etc...
* Identities, Functions, Endpoints, and Topics may all be grouped, similar to "roles" in other systems.
* Groups, Functions, Endpoints, and Topics have an owning identity or identity group.
* The owner of a group, Function, Endpoint, or Topic may grant permissions to others on their owned resources.
* When a identity calls a function through the gateway, they pass their API token along in a header
* That token is passed into the context of called functions, and is used in subsequent calls to the
  gateway. A called function assumes the identity of the caller, avoiding privilege escalation attacks by
  conflating code with owning teams in cases where you never want a user to even indirectly trigger access
  on a particular system, similar to the MLS security model. Breaking a function's security should not grant 
  the attacker the ability to reconfigure other resources owned by the function's owning identity.
* This is based on evaluations of existing multitenancy solutions in popular cloud-based systems, as well
  as ideas from [Attribute-Based Access Control](https://en.wikipedia.org/wiki/Attribute-based_access_control)
  to improve the expressivity of the rules that we allow.

#### Example Rules

* function A can call function B
* all functions in AZ 1 can call function group C in AZ 3
* a function triggered by user Eve should never (even indirectly) call our payment system functions.
* unauthenticated users should never, even indirectly, trigger data to be published in our "admin comments" topic

## Workflows

### Initial Setup

1. admin starts system
1. admin changes default admin token
1. admin creates a new identity group, currently empty, called "dev team". "dev team" is
   owned by the creator, admin.
1. admin creates a new identity, "alice" who is then assigned ownership of "dev team" and
   granted membership
1. admin grants "dev team" the freedom to create functions, topics, and endpoints

### Creating and Using Functions and Topics

1. alice is given their API token by the admin
1. using their API token, alice creates function f1 and topic t1. f1 is configured
   by alice as a producer that sends its output to t1.
1. admin creates a "business intelligence" team, adds identity "zoltan" to it, and grants
   the team the ability to create new functions.
1. zoltan creates a function, "t1 analyzer"
1. zoltan requests that alice either allow zoltan to have permission to add new
   subscribers to the topic t1, or for alice to register "t1 analyzer" as a
   subscriber to the topic t1.
1. alice adds "t1 analyzer" as a subscriber to t1.

### Creating and Using Endpoints

1. alice creates endpoint e1, and sets f1 as its backing function
1. zoltan tries to hit e1, but they did not have permission to access it, and
   the request is denied.
1. zoltan asks alice to make e1 completely open, or to grant zoltan specific
   permission to reach it.
1. alice configures e1 to be callable by the world with no auth required
1. zoltan successfully tests their function, and starts processing
   all results of f1 as identitys around the world access it.

## Usage

The identity's token is hydrated into serverless.yml through an
env var, based on the deployment environment.

The sdk passes along the token in an http header with every request to the API.

### Namespaces

```
// as admin
sdk.createNamespace("analytics")
sdk.createIdentity("hendrik")
sdk.bindToken("hendrik", "120347aea9d1f25c1ca3b4d64eb561947e8418b33d")
sdk.assignNamespace("function", "f1", "analytics") // type, object, namespace
sdk.assignNamespace("identity", "hendrik", "analytics")

// hendrik can now call f1 by passing their token in a header to the gateway
```

### Rules

```
// as admin
sdk.createIdentity("alice")
sdk.bindToken("alice", "120347aea9d1f25c1ca3b4d64eb561947e8418b33d")
sdk.grant("alice", "create", "topics")
sdk.grant("alice", "create", "functions")

// as alice, who passes the token 120347aea9d1f25c1ca3b4d64eb561947e8418b33d in a header
sdk.createFunction("f1"...)
sdk.createTopic("t1"...)

// admin creates new user
sdk.createIdentity("bob")
sdk.bindToken("bob", "a9d1f25c1ca3b4d64eb561947e")
sdk.grant("bob", "create", "functions")

// alice grants permissions on things they own to bob
sdk.grant("bob", "call", "f1")
sdk.grant("bob", "subscribe", "t1")

// bob uses their new access, passing their token along
sdk.createFunction("f2"...)
sdk.subscribe("f2", "t1")

// eve does not have an identity, or they have one without permissions
sdk.grant("eve", "ownership", "t1") // FAILS
```

## Implementation Path

the following MAY be possible by emit:

1. feature on/off switch: add a flag to start the gateway in mandatory access control mode
1. storage: identity to tokens mapping
1. storage: identity to namespaces mapping
1. api: identity management CRUD
1. api: topic, function, endpoint ownership CRUD
1. api: namespace management CRUD
1. api: thread namespace enforcement into all existing config api's
1. router: thread namespace enforcement into endpoint decisions
1. router: thread namespace enforcement into pub/sub decisions
1. encryption: add flag for symmetric encryption key to gateway
1. encryption: encrypt all keys and values in the backing database

----- EMIT -----

1. storage: identity to associated rule mapping
1. api: rule management CRUD
1. api: thread rule enforcement into all existing config api's
1. router: thread rule enforcement into endpoint decisions
1. router: thread rule enforcement into pub/sub decisions
