# Undergraduate Final Distributed System
## About
This repo contains all my notes, code, and configuration I learned or needed
for my undergraduate final about distributed system

## What I do for the final
My undergraduate final is "Multi-tenancy Implementation for Provisioning Kubernetes 
Cluster". Basically, I need to implement multi-tenancy (more than one user) on
the kubernetes cluster. There are several ways to do this, one of them is using virtual
cluster

## Work In Progress
- [ ] simple scenario, 1 cluster 1 machine
- [ ] above scenario, but a machine want to join the cluster
- [ ] create a deployment

## Progression
- Monday, 7 October 2024
  - Picked the final title

- Wednesday, 9 October 2024
  - First assignment
  - start to learn how Kubernetes work

- Wednesday, 13 November 2024
  - Learning the kubernetes the hard way: configured the cluster without minikube, kubeadm, etc.
  - starting to know how the kubernetes components work

- Wednesday, 20 November 2024
  - Using minikube again, learn about vcluster to create virtual cluster

# UPDATES
What is called multi-tenancy here is the nodes that join the cluster
will run in vm/container. Each computer may have more than one node (vm)
and each one may be part of different cluster
