# Event Gateway on Kubernetes (minikube)

To develop and deploy the `event-gateway` and all related elements locally, the easiest method includes using 
the [minikube](https://github.com/kubernetes/minikube) toolset. To get started, set up your local cluster with the 
following instructions...

## Contents
1. [Fedora/RHEL/CentOS](#fedora-rhel-centos)
1. [Debian/Ubuntu](#debian-ubuntu)
1. [MacOS](#macos)

### Fedora/RHEL/CentOS
+ Install the prerequisite packages:
  ```bash
  sudo dnf install kubernetes libvirt-daemon-kvm qemu-kvm nodejs docker
  ```

+ Ensure your user is added to the `libvirt` group for VM access. You can verify with `getent group libvirt` once done.
  ```bash
  sudo usermod -a -G libvirt $(whoami)
  ```

+ Next, add the `libvirt` group to your current user grouplist. Verify by running `id` once done.
  ```bash
  newgrp libvirt
  ```

+ Add the [docker-machine](https://github.com/docker/machine) binary to your system 
  ```bash
  curl -L https://github.com/docker/machine/releases/download/v0.15.0/docker-machine-$(uname -s)-$(uname -m) >/tmp/docker-machine && \
  chmod +x /tmp/docker-machine && \
  sudo cp /tmp/docker-machine /usr/local/bin/docker-machine
  ```

+ Add the CentOS `docker-machine` kvm driver. It's ok if you're not using CentOS as the driver should **still work**&trade;
  ```bash
  sudo curl -L https://github.com/dhiltgen/docker-machine-kvm/releases/download/v0.10.0/docker-machine-driver-kvm-centos7 > /tmp/docker-machine-driver-kvm && \
  sudo chmod +x /tmp/docker-machine-driver-kvm && \
  sudo mv /tmp/docker-machine-driver-kvm /usr/local/bin/docker-machine-driver-kvm
  ```

+ Download the minikube instance for your system 
  ```bash
  curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && \
  sudo chmod +x minikube && \
  sudo mv minikube /usr/local/bin/
  ```

+ Finally, start up your minikube service! **NOTE:** the instructions recommend using `kvm2` but please use the version that matches your system install
  ```bash
  minikube start --vm-driver kvm2
  ```

+ Once everything is running you should be able to view your running cluster status
  ```bash
  minikube status
  minikube service kubernetes-dashboard --namespace kube-system
  ```

### Debian/Ubuntu

PENDING

### MacOS

PENDING
