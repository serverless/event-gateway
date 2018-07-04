data "aws_ami" "amazon-linux" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn-ami-*-x86_64-gp2"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }
}

resource "aws_instance" "bastion" {
  count = "${var.bastion_enabled ? 1 : 0}"

  ami                    = "${data.aws_ami.amazon-linux.id}"
  instance_type          = "t2.micro"
  key_name               = "${var.ssh_key}"
  subnet_id              = "${var.bastion_subnet}"
  vpc_security_group_ids = ["${aws_security_group.bastion.id}"]

  tags = "${merge(var.tags, map("Name", var.bastion_name))}"
}

resource "aws_security_group" "bastion" {
  count = "${var.bastion_enabled ? 1 : 0}"

  name   = "eg-bastion"
  vpc_id = "${var.vpc_id}"

  ingress {
    protocol    = "tcp"
    from_port   = "22"
    to_port     = "22"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}
