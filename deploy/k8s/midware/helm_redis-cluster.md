# redis cluster安装步骤

```
helm repo add bitnami https://charts.bitnami.com/bitnami

helm search repo redis-cluster

helm upgrade --install redis-cluster bitnami/redis-cluster -n redis-cluster \
--set-string password="123456" \
--set persistence.storageClass="nfs-sc-default" \
--set cluster.externalAccess.enabled=true \
--set cluster.externalAccess.service.type=LoadBalancer \
--set "cluster.externalAccess.service.loadBalancerIP={\"192.168.171.241\", \"192.168.171.242\",\"192.168.171.243\",\"192.168.171.244\",\"192.168.171.245\",\"192.168.171.246\"}"
```

> 需要安装LoadBalancer组件，参见metallb。