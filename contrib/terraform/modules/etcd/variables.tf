variable "cluster_name" {
  default = "event-gateway"
}

variable "security_groups" {
  type = "list"
}

variable "vpc_id" {}

variable "base_domain" {}

variable "instance_count" {}

variable "ssh_key" {}

variable "subnets" {
  type = "list"
}

variable "container_linux_channel" {
  default = "stable"
}

variable "ec2_type" {}

variable "tls_enabled" {}

variable "container_image" {}

variable "root_volume_iops" {}

variable "root_volume_size" {}

variable "root_volume_type" {}

variable "tags" {
  type = "map"
}

variable "bastion_enabled" {}

variable "bastion_subnet" {
  default = 0
}

variable "bastion_name" {
  default = "eg-etcd-bastion"
}

variable "use_metadata" {
  default = false
}
