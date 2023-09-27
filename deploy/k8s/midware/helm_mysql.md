# mysql 主从安装

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami

helm search repo mysql

helm upgrade --install mysql bitnami/mysql -n mysql \
--set global.storageClass=nfs-sc-default \
--set persistence.storageClass=nfs-sc-default \
--set architecture=replication \
--set-string auth.rootPassword="123456" \
--set auth.createDatabase=false \
--set auth.replicationUser=replicator \
--set-string auth.replicationPassword="123456" \
--set primary.service.type=NodePort \
--set-string primary.service.nodePorts.mysql="3306" \
--set primary.nodeSelector."kubernetes\.io/hostname"=node-2 \
--set secondary.replicaCount=1 \
--set secondary.service.type=NodePort \
--set-string secondary.service.nodePorts.mysql="3307" \
--set secondary.nodeSelector."kubernetes\.io/hostname"=node-3

mysql -uroot -p123456 -h192.168.171.102 -P3306
mysql -uroot -p123456 -h192.168.171.103 -P3307
```
