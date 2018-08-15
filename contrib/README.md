# Event Gateway on Kubernetes 

This chart deploys the Event Gateway with etcd onto a Kubernetes cluster. Please note, the default instructions expect
an existing kubernetes cluster that supports ingress, such as GKE. If your environment doesn't have ingress support
set up, please follow the [minikube](MINIKUBE.md) instructions to set this up for your development environment.  

## Contents

1. [Quickstart](#quickstart)
    1. [Using helm](#using-helm)
    1. [Using custom resources](#using-custom-resources)
1. [Examples](#examples)
1. [Configuration](#configuration)
1. [Cleanup](#cleanup)

## Quickstart

Make sure you have helm installed on your machine and run `helm init` on your k8s cluster. This will set up the
`helm` and `tiller` functions required for easy deployment of config files to your cluster. You can follow
instructions [here](https://docs.helm.sh/using_helm/#quickstart) if you have not set this up previously.

**NOTE:** This portion of the config expects you to have a pre-existing kubernetes cluster (not minikube). For 
local development please check the [minikube](MINIKUBE.md) information below.

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

Next we'll need to collect the Event Gateway IP to use on the CLI. To do so, inspect your services as follows:

```bash
export EVENT_GATEWAY_URL=$(kubectl get ingress event-gateway-ingress -o jsonpath={.status.loadBalancer.ingress[0].ip})
```

With your environment set up, you can now jump to the [examples](#examples) section to put your `event-gateway` to use!

### Using helm

Most of the instructions for using `helm` come from the [Quickstart](#quick-start), but please note the differnce when collecting
the `config` and `events` API ports. Minikube does not ship with integrated loadbalancer options like hosted environments would
provide (e.g. Google Kubernetes Engine). As a result, though we can use the same `helm` charts as those installations, we'll
need to grab our ports from the randomly assigned `nodePort` fields before moving forward. There are numerous articles in the 
community that describe this minikube-specific behavior, but they are beyond the scope of this document 
(edit: [here](https://kubernetes.io/docs/tutorials/kubernetes-basics/expose/expose-intro/) is a bit of information on exposing services).

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
this namespace has no bearing on your Event Gateway `spaces` as outlined in the [docs](https://github.com/serverless/event-gateway/blob/master/README.md). 
If you'd like to install `etcd-operator` and `event-gateway` in another namespace, add the `--namespace <namespace>` option to 
both `helm install` commands above.

Next we'll need to collect the Event Gateway IP and ports for use on the CLI. To do so, inspect your services as follows:

```bash
export EVENT_GATEWAY_URL=$(minikube ip)
export EVENT_GATEWAY_CONFIG_API_PORT=$(kubectl get svc eg-event-gateway -o json | jq -r '.spec.ports[] | select(.name=="config") | .nodePort | tostring')
export EVENT_GATEWAY_EVENTS_API_PORT=$(kubectl get svc eg-event-gateway -o json | jq -r '.spec.ports[] | select(.name=="events") | .nodePort | tostring')
```

This should yield something like the following (your data will be dependent on your specific cluster):
```bash
$ env | grep EVENT
...
EVENT_GATEWAY_URL=192.168.42.202
EVENT_GATEWAY_EVENTS_API_PORT=31455
EVENT_GATEWAY_CONFIG_API_PORT=30523
```

With your environment set up, you can now jump to the [examples](#examples) section to put your `event-gateway` to use!

### Using custom resource definitions

PENDING

## Examples

Once you've set each of the `EVENT_GATEWAY_URL`, `EVENT_GATEWAY_CONFIG_API_PORT`, and `EVENT_GATEWAY_EVENTS_API_PORT` environment 
variables, you're set to start interacting with the `event-gateway`! 

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
  --url http://${EVENT_GATEWAY_URL}:${EVENT_GATEWAY_CONFIG_API_PORT}/v1/spaces/default/functions \
  --header 'content-type: application/json' \
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
   --url http://${EVENT_GATEWAY_URL}:${EVENT_GATEWAY_CONFIG_API_PORT}/v1/spaces/default/functions \
   --header 'content-type: application/json' \
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
  --url http://${EVENT_GATEWAY_URL}:${EVENT_GATEWAY_CONFIG_API_PORT}/v1/spaces/default/functions \
  --header 'content-type: application/json' | jq
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

**Register an event**

**Register a subscription**

**Trigger an event**

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
| `command`                   | Options to pass to `event-gateway` command   | `[-db-hosts=eg-etcd-cluster-client:2379, -log-level=debug]`|
| `etcd_cluster_name`         | Name of the etcd cluster. Passed to the `-db-hosts` option as `<etcd-cluster-name>-client`  | `eg-etcd-cluster`|

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
