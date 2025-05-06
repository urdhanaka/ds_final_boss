#!/bin/bash

LATEST_VERSION=$(curl -w '%{url_effective}' -L -s -S https://update.k3s.io/v1-release/channels/stable -o /dev/null | sed -e 's|.*/||')

install_latest_version() {
  curl -Lo /usr/local/bin/k3s https://github.com/k3s-io/k3s/releases/download/${LATEST_VERSION}/k3s; chmod a+x /usr/local/bin/k3s
}

# check if k3s binary is in PATH or the local version is not the same as the latest version
# if not, download the k3s binary
if ! command -v k3s 2>&1; then
  install_latest_version 
else
  LOCAL_K3S_VERSION=$(k3s -v | head --lines=1 | awk '{print $2}')
  if [[ ${LOCAL_K3S_VERSION} != ${LATEST_VERSION} ]]; then
    install_latest_version 
  fi
fi
