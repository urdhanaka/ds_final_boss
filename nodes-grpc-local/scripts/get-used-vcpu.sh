#!/bin/bash
vcpu_used=0
active_domain_list=$(virsh list | awk 'NR>2 {print $2}')

if [ -z "$active_domain_list" ]; then
  echo "$vcpu_used" | tr -d '\n'
else
  while read -r vm_name; do
    inside="$vm_name"
    this_domain_max_vcpu=$(virsh vcpucount --domain "$inside" --maximum | tr -d '\n');
    vcpu_used=$((vcpu_used+this_domain_max_vcpu))
  done <<< "$active_domain_list"

  echo "$vcpu_used" | tr -d '\n'
fi
