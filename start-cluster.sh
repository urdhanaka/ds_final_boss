#!/bin/bash

# make sure current cluster deleted first
kind delete cluster --name=ta && kind create cluster --config=kind-config.yaml --wait=30s

# get cluster kubernetes client certificate
kind get kubeconfig --name=ta | yq e '.users[0].user.client-certificate-data' - | base64 -d > ./kube-config/kind.csr

# copy kubernetes root certificate
podman cp ta-control-plane:/etc/kubernetes/pki/ca.crt ./kube-config/
podman cp ta-control-plane:/etc/kubernetes/pki/ca.key ./kube-config/

# create users private key
openssl genrsa -out ./kube-config/user-a.key 2048
openssl genrsa -out ./kube-config/user-b.key 2048

openssl req -new -key ./kube-config/user-a.key -out ./kube-config/user-a.csr -subj "/CN=user-a/O=tenant-a"
openssl req -new -key ./kube-config/user-b.key -out ./kube-config/user-b.csr -subj "/CN=user-b/O=tenant-b"

openssl x509 -req -in ./kube-config/user-a.csr -CA ./kube-config/ca.crt -CAkey ./kube-config/ca.key -CAcreateserial -out ./kube-config/user-a.crt
openssl x509 -req -in ./kube-config/user-b.csr -CA ./kube-config/ca.crt -CAkey ./kube-config/ca.key -CAcreateserial -out ./kube-config/user-b.crt

# create user kubeconfig
cat > ./kube-config/user-a-config.yaml << EOF
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: $(cat ./kube-config/ca.crt | base64 | tr -d '\n')
    server: $(kubectl cluster-info | grep "control plane" | awk -F" " {'print $7'})
  name: kind-ta
contexts:
- context:
    cluster: kind-ta
    user: user-a
  name: kind-ta
current-context: kind-ta
kind: Config
preferences: {}
users:
- name: user-a
  user:
    client-certificate-data: $(cat ./kube-config/user-a.crt | base64 | tr -d '\n')
    client-key-data: $(cat ./kube-config/user-a.key | base64 | tr -d '\n')
EOF

cat > ./kube-config/user-b-config.yaml << EOF
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: $(cat ./kube-config/ca.crt | base64 | tr -d '\n')
    server: $(kubectl cluster-info | grep "control plane" | awk -F" " {'print $7'})
  name: kind-ta
contexts:
- context:
    cluster: kind-ta
    user: user-b
  name: kind-ta
current-context: kind-ta
kind: Config
preferences: {}
users:
- name: user-b
  user:
    client-certificate-data: $(cat ./kube-config/user-b.crt | base64 | tr -d '\n')
    client-key-data: $(cat ./kube-config/user-b.key | base64 | tr -d '\n')
EOF

# create basic rbac configuration
cat << EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: user-namespace-auth
rules:
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["get", "watch", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: user-a-namespace-auth
subjects:
- kind: User
  name: user-a
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: user-namespace-auth
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: user-b-namespace-auth
subjects:
- kind: User
  name: user-b
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: user-namespace-auth
  apiGroup: rbac.authorization.k8s.io
EOF

kubectl create namespace tenant-user-a
kubectl create namespace tenant-user-b

cat << EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: tenant-user-a
  name: tenant-a-full-access
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["get", "watch", "list", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: tenant-user-b
  name: tenant-b-full-access
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["get", "watch", "list", "create", "update", "patch", "delete"]
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: full-access-binding-a
  namespace: tenant-user-a
subjects:
- kind: User
  name: user-a
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: tenant-a-full-access
  apiGroup: rbac.authorization.k8s.io
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: full-access-binding-b
  namespace: tenant-user-b
subjects:
- kind: User
  name: user-b
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: tenant-b-full-access
  apiGroup: rbac.authorization.k8s.io
EOF
