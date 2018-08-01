# Event Gateway Helm chart

This chart deploys the Event Gateway with etcd onto a Kubernetes cluster.

## Installation

Make sure you have helm installed on you machine and run `helm init` on your K8s cluster.

From the `event-gateway/contrib/helm` folder:

First, install etcd operator:
```
helm install stable/etcd-operator --name ego
```

Then, install the Event Gateway:
```
helm install event-gateway --name eg
```

Note: to deploy the stack to a namespace other than the default, add `--namespace` option to both `helm install` commands.

To get the Event Gateway load balancer IP:
```
kubectl get svc
```

To delete the Event Gatway and etcd:
```
helm delete eg
helm delete ego
```

## Configuration

| Parameter                   | Description                                  | Default                    |
|-----------------------------|----------------------------------------------|----------------------------|
| `images.repository`         | Event Gateway image                          | `serverless/event-gateway` |
| `images.tag`                | Event Gateway image tag                      | `0.9.0`                    |
| `replicaCount`              | Number of containers                         | `3`                        |
| `service.type`              | Type of Kubernetes service                   | `LoadBalancer`             |
| `service.config.port`       | Config API port number                       | `4001`                     |
| `service.events.port`       | Events API port number                       | `4000`                     |
| `resources.limits.cpu`      | CPU resource limits                          | `200m`                     |
| `resources.limits.memory`   | Memory resource limits                       | `256Mi`                    |
| `resources.requests.cpu`    | CPU resource requests                        | `200m`                     |
| `resources.requests.memory` | Memory resource requests                     | `256Mi`                    |
| `command`                   | Options to pass to `event-gateway` command   | `[-db-hosts=eg-etcd-cluster-client:2379, -log-level=debug]`|
| `etcd_cluster_name`         | Name of the etcd cluster. Must be passed to the `-db-host` option as `<etcd-cluster-name>-client`  | `eg-etcd-cluster`|