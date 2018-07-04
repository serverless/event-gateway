locals {
  scheme = "${var.tls_enabled ? "https" : "http"}"

  // see https://github.com/hashicorp/terraform/issues/9858
  etcd_initial_cluster_list = "${concat(data.template_file.etcd_hostname_list.*.rendered, list("dummy"))}"

  metadata_env = "EnvironmentFile=/run/metadata/coreos"

  metadata_deps = <<EOF
Requires=coreos-metadata.service
After=coreos-metadata.service
EOF

  cert_options = <<EOF
--cert-file=/etc/ssl/etcd/server.crt \
  --client-cert-auth=true \
  --key-file=/etc/ssl/etcd/server.key \
  --peer-cert-file=/etc/ssl/etcd/peer.crt \
  --peer-key-file=/etc/ssl/etcd/peer.key \
  --peer-trusted-ca-file=/etc/ssl/etcd/ca.crt \
  --peer-client-cert-auth=true \
  --trusted-ca-file=/etc/ssl/etcd/ca.crtEOF
}

data "template_file" "etcd_hostname_list" {
  count = "${var.instance_count}"

  template = "${var.cluster_name}-etcd-${count.index}.${var.base_domain}"
}

data "template_file" "etcd_names" {
  count    = "${var.instance_count}"
  template = "${var.cluster_name}-etcd-${count.index}${var.base_domain == "" ? "" : ".${var.base_domain}"}"
}

data "template_file" "advertise_client_urls" {
  count    = "${var.instance_count}"
  template = "${local.scheme}://${data.template_file.etcd_hostname_list.*.rendered[count.index]}:2379"
}

data "template_file" "initial_advertise_peer_urls" {
  count    = "${var.instance_count}"
  template = "${local.scheme}://${data.template_file.etcd_hostname_list.*.rendered[count.index]}:2380"
}

data "template_file" "initial_cluster" {
  count    = "${length(data.template_file.etcd_hostname_list.*.rendered) > 0 ? var.instance_count : 0}"
  template = "${data.template_file.etcd_names.*.rendered[count.index]}=${local.scheme}://${local.etcd_initial_cluster_list[count.index]}:2380"
}

data "template_file" "etcd" {
  count    = "${var.instance_count}"
  template = "${file("${path.module}/resources/dropins/40-etcd-cluster.conf")}"

  vars = {
    advertise_client_urls       = "${data.template_file.advertise_client_urls.*.rendered[count.index]}"
    cert_options                = "${var.tls_enabled ? local.cert_options : ""}"
    container_image             = "${var.container_image}"
    initial_advertise_peer_urls = "${data.template_file.initial_advertise_peer_urls.*.rendered[count.index]}"
    initial_cluster             = "${length(data.template_file.etcd_hostname_list.*.rendered) > 0 ? format("--initial-cluster=%s", join(",", data.template_file.initial_cluster.*.rendered)) : ""}"
    metadata_deps               = "${var.use_metadata ? local.metadata_deps : ""}"
    metadata_env                = "${var.use_metadata ? local.metadata_env : ""}"
    name                        = "${data.template_file.etcd_names.*.rendered[count.index]}"
    scheme                      = "${local.scheme}"
  }
}

data "ignition_systemd_unit" "etcd" {
  count   = "${var.instance_count}"
  name    = "etcd-member.service"
  enabled = true

  dropin = [
    {
      name    = "40-etcd-cluster.conf"
      content = "${data.template_file.etcd.*.rendered[count.index]}"
    },
  ]
}
