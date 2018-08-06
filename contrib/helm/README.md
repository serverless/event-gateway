# Event Gateway Helm chart

This chart deploys the Event Gateway with etcd onto a Kubernetes cluster. Please note, the default instructions expect
an existing kubernetes cluster that supports loadbalancers, such as GKE. If your environment doesn't have a loadbalancer
set up, please follow the minikube instructions below to retrieve the `event_gateway` information.

## Contents

1. [Quickstart](#quickstart)
1. [Minikube or local clusters](#minikube-or-local-clusters)
    1. [Setting up an Ingress](#setting-up-an-ingress)
1. [Configuration](#configuration)

### Quickstart

Make sure you have helm installed on you machine and run `helm init` on your k8s cluster. This will set up the
`helm` and `tiller` functions required for easy deployment of config files to your cluster. You can follow
instructions [here](https://docs.helm.sh/using_helm/#quickstart) if you have not set this up previously.

**NOTE:** This portion of the config expects you to have a pre-existing kubernetes cluster (not minikube). For 
local development please check the [minikube](#minikube-or-local-clusters) information below.

Once installed, navigate to the `event-gateway/contrib/helm` folder and install the following components:

**etcd-operator**
```
helm install stable/etcd-operator --name ego [--namespace <namespace>]
```

**event-gateway**
```
helm install event-gateway --name eg [--namespace <namespace>]
```

This will install each of the `etcd-operator` and `event-gateway` into the `default` namespace in kubernetes. Please note,
this namespace has no bearing on your Event Gateway `spaces` as outlined in the [docs](https://github.com/serverless/event-gateway/blob/master/README.md). If you'd like to install `etcd-operator` and `event-gateway` in another namespace, add the `--namespace <namespace>` option to both `helm install` commands above.

Next we'll need to collect the Event Gateway IP and ports for use on the CLI. To do so, inspect your services as follows:

```
export EVENT_GATEWAY_URL=$(kubectl get svc event-gateway -o jsonpath={.status.loadBalancer.ingress[0].ip})
export EVENT_GATEWAY_CONFIG_API_PORT=4001
export EVENT_GATEWAY_EVENTS_API_PORT=4000
```

To get the Event Gateway load balancer IP:
```
kubectl get svc
```

To delete the Event Gatway and etcd:
```
helm delete eg
helm delete ego
```

### Minikube or local clusters

#### Setting up an Ingress

### Configuration

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
