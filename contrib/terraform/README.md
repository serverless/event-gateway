# Event Gateway Terraform module

This module creates Event Gateway running on ECS Fargate with a standalone etcd cluster.

The module is an extract form the [Tectonic Installer repository](https://github.com/coreos/tectonic-installer).

## Usage

```hcl
module "event-gateway" {
  source = "github.com/serverless/event-gateway//contrib/terraform/modules/event-gateway"

  aws_region = "us-east-1"
  command_list = ["-db-hosts", "event-gateway-etcd-0.etcd:2379,event-gateway-etcd-1.etcd:2379,event-gateway-etcd-2.etcd:2379", "-log-level", "debug"]
  tags = {
    Application = "event-gateway"
  }
}

output "config_url" {
  value = "${module.event-gateway.config_url}"
}

output "events_url" {
  value = "${module.event-gateway.events_url}"
}
```

## Debugging etcd

It's possible to enable SSH access via bastion instance, by adding parameters:

```
bastion_enabled = true
ssh_key      = "eg-key"
```

Bastion IP can be distplayed by adding output:

```
output "bastion_ip" {
  value = "${module.event-gateway.bastion_ip}"
}
```

To connect to one of the etcd cluster hosts, run:

```bash
ssh -J ec2-user@<bastion_ip> core@<etcd_host_private_ip>
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| aws_region | AWS region for the stack | string | - | yes |
| bastion_enabled | Set to true enables SSH access to etcd nodes in the private subnet | string | `false` | no |
| command_list | List of parameters for the `event-gateway` command | list | `["-log-level", "debug"]` | no |
| config_alb_name | Config ALB name | string | `alb-config` | no |
| config_port | Port number of the Event Gateway Config API | string | `4001` | no |
| eg_image | Event Gateway docker image | string | `serverless/event-gateway:latest` | no |
| eg_vpc_name | Event Gateway VPC name | string | `eg-vpc` | no |
| etcd_base_domain | Name of the base domain for the etcd cluster | string | `etcd` | no |
| etcd_image | etcd Docker image | string | `quay.io/coreos/etcd:v3.1.8` | no |
| etcd_instance_count | Number of nodes in the etcd cluster | string | `3` | no |
| etcd_instance_type | Etcd node type | string | `t2.micro` | no |
| etcd_root_volume_iops | Number of IOPS of the etcd cluster volumes | string | `100` | no |
| etcd_root_volume_size | Size of the etcd cluster volumes (in GiB) | string | `30` | no |
| etcd_root_volume_type | Type of the etcd cluster volumes | string | `gp2` | no |
| etcd_ssh_key | (optional) Name of the preexisting SSH key | string | `` | no |
| etcd_tls_enabled | Enable TLS for the etcd cluster | string | `false` | no |
| events_alb_name | Events ALB name | string | `alb-events` | no |
| events_port | Event Gateway Events API port number | string | `4000` | no |
| fargate_cpu | Fargate instance CPU units | string | `256` | no |
| fargate_memory | Fargate instance memory | string | `512` | no |
| tags | Additional tags | map | `<map>` | no |
| task_count | Number of Event Gateway Fargate tasks | string | `3` | no |

## Outputs

| Name | Description |
|------|-------------|
| bastion_ip | Public IP of etcd bastion instance |
| config_url | Event Gateway Config API URL |
| events_url | Event Gateway Events API URL |
