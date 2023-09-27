# redis cluster安装步骤

```
helm repo add bitnami https://charts.bitnami.com/bitnami

helm search repo etcd

helm upgrade --install etcd-cluster  bitnami/etcd -n etcd-cluster \
--set replicaCount=3 \
--set-string auth.rbac.rootPassword="123456"

```