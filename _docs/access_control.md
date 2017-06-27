# Access Control Spec

## Goals

* enable organizations to share resources in a familiar, easy way
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

* Rules are defined as a tuple of (Subject, Action, Object, Environment)
* Subject may be a user or a group of users
* Object may be a user, function, endpoint, topic, or a group of entities from one type of object
* Environment may relate to a logical namespace, datacenter, geographical region or time
* Users, Functions, Endpoints, and Topics may all be grouped, similar to "roles" in other systems.
* Groups may only be comprised of a single type of entity, no mixing users and functions, for example.
* Users, Functions, Endpoints, Topics, and groups of them may be created, deleted, granted
  permissions, and revoked of their permissions.
* Groups, Functions, Endpoints, and Topics have an owning user or user group.
* The owner of a group, Function, Endpoint, or Topic may grant permissions to others on their owned resources.
* Users have a revokable API token, which may be retrieved by authenticating with the gateway, or
  distributed by an administrator
* When a user calls a function through the gateway, they pass their API token along in a header
* That token is passed into the context of called functions, and is used in subsequent calls to the
  gateway. A called function assumes the identity of the caller, avoiding privilege escalation attacks by
  conflating code with owning teams. Breaking a function's security should not grant the attacker the
  ability to reconfigure other resources owned by the same team.
* This is based on evaluations of existing multitenancy solutions in popular cloud-based systems, as well
  as ideas from [Attribute-Based Access Control](https://en.wikipedia.org/wiki/Attribute-based_access_control)
  to improve the expressivity of the rules that we allow.

## Workflows

### Initial Setup

1. admin logs in, changes default login credentials
1. admin creates a new user group, currently empty, called "dev team". "dev team" is
   owned by the creator, admin.
1. admin creates a new user, "alice" who is then assigned ownership of "dev team" and
   granted membership
1. admin grants "dev team" the freedom to create functions, topics, and endpoints

### Creating and Using Functions and Topics

1. alice logs in and retrives their API token
1. using their API token, alice creates function f1 and topic t1. f1 is configured 
   by alice as a producer that sends its output to t1.
1. admin creates a "business intelligence" team, adds user "zoltan" to it, and grants
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
   all results of f1 as users around the world access it.
