# Event Gateway on Kubernetes 

This chart deploys the Event Gateway with etcd onto a Kubernetes cluster. Please note, the default instructions expect
an existing kubernetes cluster that supports ingress, such as GKE. If your environment doesn't have ingress support
set up, please follow the [minikube](MINIKUBE.md) instructions to set this up for your development environment.  

## Contents

1. [Quickstart](#quickstart)
1. [Examples](#examples)
    1. [Register a function](#register-a-function)
    1. [Query all function](#query-all-function)
    1. [Register an event](#register-an-event)
    1. [Query all events](#query-all-events)
    1. [Register a subscription](#register-a-subscription)
    1. [Query all subscriptions](#query-all-subscriptions)
    1. [Trigger an event](#trigger-an-event)
1. [Configuration](#configuration)
1. [Cleanup](#cleanup)

## Quickstart

Make sure you have helm installed on your machine and run `helm init` on your k8s cluster. This will set up the
`helm` and `tiller` functions required for easy deployment of config files to your cluster. You can follow
instructions [here](https://docs.helm.sh/using_helm/#quickstart) if you have not set this up previously.

**NOTE:** This portion of the config expects you to have a pre-existing kubernetes cluster (not minikube). For 
local development please check the [minikube](MINIKUBE.md) information.

Once installed, navigate to the `event-gateway/contrib/helm` folder and install the following components:

**etcd-operator**
```bash
helm install stable/etcd-operator --name ego [--namespace <namespace>]
```

**event-gateway**
```bash
helm install event-gateway --name eg [--namespace <namespace>]
```

This will install each of the `etcd-operator` and `event-gateway` into the `default` namespace in kubernetes. Please note,
this namespace has no bearing on your Event Gateway `spaces` as outlined in the [docs](https://github.com/serverless/event-gateway/blob/master/README.md). If you'd like
to install `etcd-operator` and `event-gateway` in another namespace, add the `--namespace <namespace>` option to
 both `helm install` commands above.

Next we'll need to collect the Event Gateway IP to use on the CLI. We have a couple of options available to reference the
internal services of the kubernetes cluster exposed via Ingress:

1. add Ingress IP to /etc/hosts (recommended)

   This method enables us to reference the `event-gateway` from the hostname we configured in the Ingress module. This document 
   assumes the name to be `eventgateway.minikube` so please update the instructions to your naming convention if you need.
  
   ```bash
     echo "$(kubectl get ingress event-gateway-ingress -o jsonpath={.status.loadBalancer.ingress[0].ip}) eventgateway.minikube" | sudo tee -a "/etc/hosts"
   ```

1. use Ingress IP and pass header to request

   With this method we access the `event-gateway` using the IP of the Ingress directly. Since the Ingress was configured to 
   receive all connections from the `eventgateway.minikube` host, you'll need to pass this as a header value to the request.
   
   ```bash
   export EVENT_GATEWAY_URL=$(kubectl get ingress event-gateway-ingress -o jsonpath={.status.loadBalancer.ingress[0].ip})
   curl --request GET \
        --url http://{EVENT_GATEWAY_URL}/v1/metrics \
        --header 'content-type: application/json' \
        --header 'host: eventgateway.minikube'
   ```

With your environment set up, you can now jump to the [examples](#examples) section to put your `event-gateway` to use!

## Examples

Once you've set the `EVENT_GATEWAY_URL` environment variable, you're set to start interacting with the `event-gateway`! 

**NOTE:** the events and configuration API ports are abstracted away from us via the kubernetes Ingress. The path-based 
routing will ensure the request goes to the proper service managed by the cluster. 

**DOUBLENOTE:** if you did not want to use an environment variable for connecting to the `event-gateway`, you can use
the host of your Ingress by adding to `/etc/hosts`. Please check the [Quickstart](#quickstart) for reference. 

**TRIPLENOTE:** the examples below all assume the `default` namespace for the `event-gateway`. If you've updated or changed
this on your end, please don't forget to update the queries accordingly.

#### Register a function

Define the function registration payload, using **AWS** as an example:

```bash
cat > function.json <<EOF
{
    "functionId": "echo",
    "type": "awslambda",
    "provider": {
        "arn": "arn:aws:lambda:us-east-1:123456789012:function:event-gateway-tests-dev-echo",
        "region": "us-east-1",
        "awsAccessKeyID": "AAAAAAAAAAAAAAAAAAAA",
        "awsSecretAccessKey": "AAAAaBcDeFgHiJqLmNoPqRsTuVwXyz0123456789"
    }
}
EOF
```

Then call the registration endpoint with your json payload:

```bash
curl --request POST \
  --url http://${EVENT_GATEWAY_URL}/v1/spaces/default/functions \
  --header 'content-type: application/json' \
  --header 'host: eventgateway.minikube' \
  --data @function.json
```

And the corresponsing reply (if successful) should read something like the following:

```bash
{
	"space": "default",
	"functionId": "echo",
	"type": "awslambda",
	"provider": {
		"arn": "arn:aws:lambda:us-east-1:123456789012:function:event-gateway-tests-dev-echo",
		"region": "us-east-1",
		"awsAccessKeyId": "AAAAAAAAAAAAAAAAAAAA",
		"awsSecretAccessKey": "AAAAaBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789"
	}
}
```

**NOTE:** if you try to overwrite an existing function, you will receive an error! To replace an existing function
you will have to delete it first, then register the function once more. For example, trying to re-register the `echo`
function will yield:

```bash
curl --request POST \
   --url http://${EVENT_GATEWAY_URL}/v1/spaces/default/functions \
   --header 'content-type: application/json' \
   --header 'host: eventgateway.minikube' \
   --data @function.json

{
    "errors": [{
        "message": "Function \"echo\" already registered."
    }]
}
```

#### Query all functions

To check for registered functions, query the `config` API with the `GET` request:

```bash
curl --request GET \
  --url http://${EVENT_GATEWAY_URL}/v1/spaces/default/functions \
  --header 'content-type: application/json' \
  --header 'host: eventgateway.minikube'  | jq
```

You should see the functions list return your defined set of functions across all vendors.

```bash
{
  "functions": [
    {
      "space": "default",
      "functionId": "echo",
      "type": "awslambda",
      "provider": {
        "arn": "arn:aws:lambda:us-east-1:123456789012:function:event-gateway-tests-dev-echo",
        "region": "us-east-1",
        "awsAccessKeyId": "AAAAAAAAAAAAAAAAAAAA",
        "awsSecretAccessKey": "AAAAaBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789"
      }
    }
  ]
}
```

#### Register an event

To register an event, make sure to `POST` the event name to the `event-gateway`.

```bash
curl --request POST \
  --url http://eventgateway.minikube/v1/spaces/default/eventtypes \
  --header 'content-type: application/json' \
  --data '{ "name": "eventgateway.function.invoked" }'
```

The reply should look something like the following:

```bash
{ 
  "space": "default",
  "name": "eventgateway.function.invoked"
}
```

#### Query all events

```bash
curl --request GET \
  --url http://eventgateway.minikube/v1/spaces/default/eventtypes \
  --header 'content-type: application/json'
```

Your registered events reply should look as follows:

```bash
{
  "eventTypes": [
    {
      "space": "default",
      "name": "eventgateway.function.invoked"
    }
  ]
}
```

#### Register a subscription

To register subscriptions to one of your registered event types, make sure to specify the `eventType` in 
the JSON POST payload.

```bash
curl --request POST \
  --url http://eventgateway.minikube/v1/spaces/default/subscriptions \
  --header 'content-type: application/json' \
  --data '{
    "type": "async",
    "eventType": "eventgateway.function.invoked",
    "functionId": "echo",
    "path": "/",
    "method": "POST"
}'
```

Your reply payload should include the `subscriptionId` for your new subscription:

```bash
{
  "space": "default",
  "subscriptionId": "YXN5bmMsZXZlbnRnYXRld2F5LmZ1bmN0aW9uLmludm9rZWQsZWNobywlMkYsUE9TVA",
  "type": "async",
  "eventType": "eventgateway.function.invoked",
  "functionId": "echo",
  "path": "/",
  "method": "POST"
}
```

#### Query all subscriptions

To list our your current subscrptions, you can do the following:

```bash
curl --request GET \
  --url http://eventgateway.minikube/v1/spaces/default/subscriptions \
  --header 'content-type: application/json' \
```

The output should list each of the registered subscriptions:

```bash
{
  "subscriptions": [
    {
      "space": "default",
      "subscriptionId": "YXN5bmMsZXZlbnRnYXRld2F5LmZ1bmN0aW9uLmludm9rZWQsZWNobywlMkYsUE9TVA",
      "type": "async",
      "eventType": "eventgateway.function.invoked",
      "functionId": "echo",
      "path": "/",
      "method": "POST"
    }
  ]
}
```

#### Trigger an event

In order to trigger a registered function, call the `event-gateway` URL with the proper `functionId`. Following
our example from earlier, we've registered a `GET` function with `functionId` set to `echo`. To trigger this function,
we would:

```bash
curl --request GET \
  --url http://eventgateway.minikube/echo \
  --header 'content-type: application/json'
```

**NOTE**: as mentioned earlier, the `events` service is handled by the path-routing service of the kubernetes Ingress. Any path
that's prepended with `/v1` will ultimately route to the `config` service, while other paths default to the `events` service. 

## Configuration

| Parameter                   | Description                                  | Default                    |
|-----------------------------|----------------------------------------------|----------------------------|
| `images.repository`         | Event Gateway image                          | `serverless/event-gateway` |
| `images.tag`                | Event Gateway image tag                      | `0.9.0`                    |
| `replicaCount`              | Number of containers                         | `3`                        |
| `service.type`              | Type of Kubernetes service                   | `LoadBalancer`             |
| `service.annotations`       | Custom annotations for the service           | `[]`                       |
| `service.config.port`       | Config API port number                       | `4001`                     |
| `service.events.port`       | Events API port number                       | `4000`                     |
| `resources.limits.cpu`      | CPU resource limits                          | `200m`                     |
| `resources.limits.memory`   | Memory resource limits                       | `256Mi`                    |
| `resources.requests.cpu`    | CPU resource requests                        | `200m`                     |
| `resources.requests.memory` | Memory resource requests                     | `256Mi`                    |
| `command`                   | Options to pass to `event-gateway` command   | `[--db-hosts=eg-etcd-cluster-client:2379, --log-level=debug]`|
| `etcd_cluster_name`         | Name of the etcd cluster. Passed to the `--db-hosts` option as `<etcd-cluster-name>-client`  | `eg-etcd-cluster`|

The service annotations can be used to set any annotations required by your platform, for example, if
you update your values.yml with:

```
-  annotations: []
+  annotations:
+    - "service.beta.kubernetes.io/aws-load-balancer-internal: 0.0.0.0/0"
+    - "foo: bar"
```

then the service will be annotated as shown:

```
$ helm install event-gateway --debug --dry-run | grep "kind: Service" -A5
kind: Service
metadata:
  name: rafting-umbrellabird-event-gateway
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-internal: 0.0.0.0/0
    foo: bar
```

## Cleanup

When you'd like to clean up the deployments, it's easy to remove services using helm: 

```bash
helm delete --purge eg
helm delete --purge ego
```
