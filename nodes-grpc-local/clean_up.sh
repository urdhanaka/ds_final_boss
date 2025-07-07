#!/bin/bash

virsh list --all | grep running | awk '{ print $2}' | while read DOMAIN; do
  virsh shutdown $DOMAIN
done
