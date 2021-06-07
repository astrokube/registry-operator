#!/bin/bash

function install_kubebuilder() {
  curl -L https://github.com/kubernetes-sigs/kubebuilder/releases/download/v3.0.0/kubebuilder_linux_amd64 >/usr/local/bin/kubebuilder
}

function install_docker() {
  wget -q https://download.docker.com/linux/static/stable/x86_64/docker-20.10.6.tgz -O /tmp/docker.tar.gz; \
  tar -xzf /tmp/docker.tar.gz -C /tmp/
  cp /tmp/docker/docker* /usr/local/bin
  chmod +x /usr/local/bin/docker*
  rm -rf /tmp/docker
}
