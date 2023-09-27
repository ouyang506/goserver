start /b cmd /c .\\etcd\\etcd-v3.5.1\\etcd.exe --name infra1  ^
  --data-dir .\\data\\etcd\\infra1 ^
  --initial-advertise-peer-urls http://127.0.0.1:2381 ^
  --listen-peer-urls http://127.0.0.1:2381 ^
  --listen-client-urls http://127.0.0.1:2378 ^
  --advertise-client-urls http://127.0.0.1:2378 ^
  --initial-cluster-token etcd-cluster-token ^
  --initial-cluster infra1=http://127.0.0.1:2381,infra2=http://127.0.0.1:2382,infra3=http://127.0.0.1:2383 ^
  --initial-cluster-state new
  
start /b cmd /c .\\etcd\\etcd-v3.5.1\\etcd.exe --name infra2 ^
  --data-dir .\\data\\etcd\\infra2 ^
  --initial-advertise-peer-urls http://127.0.0.1:2382 ^
  --listen-peer-urls http://127.0.0.1:2382 ^
  --listen-client-urls http://127.0.0.1:2379 ^
  --advertise-client-urls http://127.0.0.1:2379 ^
  --initial-cluster-token etcd-cluster-token ^
  --initial-cluster infra1=http://127.0.0.1:2381,infra2=http://127.0.0.1:2382,infra3=http://127.0.0.1:2383 ^
  --initial-cluster-state new
  
start /b cmd /c .\\etcd\\etcd-v3.5.1\\etcd.exe --name infra3 ^
  --data-dir .\\data\\etcd\\infra3 ^
  --initial-advertise-peer-urls http://127.0.0.1:2383 ^
  --listen-peer-urls http://127.0.0.1:2383 ^
  --listen-client-urls http://127.0.0.1:2380 ^
  --advertise-client-urls http://127.0.0.1:2380 ^
  --initial-cluster-token etcd-cluster-token ^
  --initial-cluster infra1=http://127.0.0.1:2381,infra2=http://127.0.0.1:2382,infra3=http://127.0.0.1:2383 ^
  --initial-cluster-state new
  
cd .\\data\\redis
start /b cmd /c ..\\..\\redis\\Redis-x64-3.0.504\\redis-server.exe
cd ..\\..\\