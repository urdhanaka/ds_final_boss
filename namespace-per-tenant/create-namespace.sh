#!/bin/bash
set -e

LAB_NAME=$1
FILENAME_TEMPLATE="create-namespace-"

cat > $FILENAME_TEMPLATE$LAB_NAME.yaml <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: $LAB_NAME
  labels:
    name: $LAB_NAME
EOF

kubectl create -f $FILENAME_TEMPLATE$LAB_NAME.yaml
