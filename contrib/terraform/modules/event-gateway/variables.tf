variable "aws_region" {
  description = "AWS region for the stack"
}

variable "command_list" {
  description = "List of parameters for the `event-gateway` command"
  type        = "list"
  default     = ["-log-level", "debug"]
}

variable "eg_image" {
  description = "Event Gateway docker image"
  default     = "serverless/event-gateway:latest"
}

variable "events_port" {
  description = "Event Gateway Events API port number"
  default     = 4000
}

variable "config_port" {
  description = "Port number of the Event Gateway Config API"
  default     = 4001
}

variable "task_count" {
  description = "Number of Event Gateway Fargate tasks"
  default     = 3
}

variable "fargate_cpu" {
  description = "Fargate instance CPU units"
  default     = 256
}

variable "fargate_memory" {
  description = "Fargate instance memory"
  default     = 512
}

variable "etcd_instance_type" {
  description = "Etcd node type"
  default     = "t2.micro"
}

variable "tags" {
  description = "Additional tags"
  type        = "map"
  default     = {}
}

variable "events_alb_name" {
  description = "Events ALB name"
  default     = "alb-events"
}

variable "config_alb_name" {
  description = "Config ALB name"
  default     = "alb-config"
}

variable "eg_vpc_name" {
  description = "Event Gateway VPC name"
  default     = "eg-vpc"
}

variable "bastion_enabled" {
  description = "Set to true enables SSH access to etcd nodes in the private subnet"
  default     = false
}

variable "etcd_root_volume_iops" {
  description = "Number of IOPS of the etcd cluster volumes"
  default     = "100"
}

variable "etcd_root_volume_size" {
  description = "Size of the etcd cluster volumes (in GiB)"
  default     = "30"
}

variable "etcd_root_volume_type" {
  description = "Type of the etcd cluster volumes"
  default     = "gp2"
}

variable "etcd_ssh_key" {
  description = "(optional) Name of the preexisting SSH key"
  default     = ""
}

variable "etcd_tls_enabled" {
  description = "Enable TLS for the etcd cluster"
  default     = false
}

variable "etcd_image" {
  description = "etcd Docker image"
  default     = "quay.io/coreos/etcd:v3.1.8"
}

variable "etcd_instance_count" {
  description = "Number of nodes in the etcd cluster"
  default     = 3
}

variable "etcd_base_domain" {
  description = "Name of the base domain for the etcd cluster"
  default     = "etcd"
}
