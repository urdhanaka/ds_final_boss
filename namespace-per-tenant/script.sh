#!/bin/bash

CLUSTER_NAME=$1

# consider that no k3s/k3d/minikube is running
# for this, we are using k3d
# with 2 worker nodes
k3d cluster create $CLUSTER_NAME --agents 2
