#!/bin/bash
set -e

NAMESPACE_NAME=$1

# check if parameter is specified
if [ -z "$NAMESPACE_NAME" ]; then
  echo "Please specify the namespace name"
  exit 1
fi

kubectl delete namespace $NAMESPACE_NAME
