locals {
  etcd_crt_id_list = [
    "${data.ignition_file.etcd_ca.*.id}",
    "${data.ignition_file.etcd_client_key.*.id}",
    "${data.ignition_file.etcd_client_crt.*.id}",
    "${data.ignition_file.etcd_server_key.*.id}",
    "${data.ignition_file.etcd_server_crt.*.id}",
    "${data.ignition_file.etcd_peer_key.*.id}",
    "${data.ignition_file.etcd_peer_crt.*.id}",
  ]
}

data "ignition_file" "etcd_ca" {
  path       = "/etc/ssl/etcd/ca.crt"
  mode       = 0644
  uid        = 232
  gid        = 232
  filesystem = "root"

  content {
    content = "${module.etcd_certs.etcd_ca_crt_pem}"
  }
}

data "ignition_file" "etcd_client_key" {
  path       = "/etc/ssl/etcd/client.key"
  mode       = 0400
  uid        = 0
  gid        = 0
  filesystem = "root"

  content {
    content = "${module.etcd_certs.etcd_client_key_pem}"
  }
}

data "ignition_file" "etcd_client_crt" {
  path       = "/etc/ssl/etcd/client.crt"
  mode       = 0400
  uid        = 0
  gid        = 0
  filesystem = "root"

  content {
    content = "${module.etcd_certs.etcd_client_crt_pem}"
  }
}

data "ignition_file" "etcd_server_key" {
  path       = "/etc/ssl/etcd/server.key"
  mode       = 0400
  uid        = 232
  gid        = 232
  filesystem = "root"

  content {
    content = "${module.etcd_certs.etcd_server_key_pem}"
  }
}

data "ignition_file" "etcd_server_crt" {
  path       = "/etc/ssl/etcd/server.crt"
  mode       = 0400
  uid        = 232
  gid        = 232
  filesystem = "root"

  content {
    content = "${module.etcd_certs.etcd_server_crt_pem}"
  }
}

data "ignition_file" "etcd_peer_key" {
  path       = "/etc/ssl/etcd/peer.key"
  mode       = 0400
  uid        = 232
  gid        = 232
  filesystem = "root"

  content {
    content = "${module.etcd_certs.etcd_peer_key_pem}"
  }
}

data "ignition_file" "etcd_peer_crt" {
  path       = "/etc/ssl/etcd/peer.crt"
  mode       = 0400
  uid        = 232
  gid        = 232
  filesystem = "root"

  content {
    content = "${module.etcd_certs.etcd_peer_crt_pem}"
  }
}
