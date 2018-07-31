output "bastion_ip" {
  value = "${element(concat(aws_instance.bastion.*.public_ip, list("")), 0)}"
}
