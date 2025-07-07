#!/bin/bash

virsh list --all | grep running | awk '{ print $2}' | while read DOMAIN; do
  virsh shutdown $DOMAIN
done

rm -rf /var/lib/libvirt/k3s-virt/*
