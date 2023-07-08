start /b cmd /c .\\etcd-v3.5.1\\etcd.exe --name infra1 --initial-advertise-peer-urls http://127.0.0.1:2381 ^
  --listen-peer-urls http://127.0.0.1:2381 ^
  --listen-client-urls http://127.0.0.1:2378 ^
  --advertise-client-urls http://127.0.0.1:2378 ^
  --initial-cluster-token etcd-cluster-1 ^
  --initial-cluster infra1=http://127.0.0.1:2381,infra2=http://127.0.0.1:2382,infra3=http://127.0.0.1:2383 ^
  --initial-cluster-state new
  
start /b cmd /c .\\etcd-v3.5.1\\etcd.exe --name infra2 --initial-advertise-peer-urls http://127.0.0.1:2382 ^
  --listen-peer-urls http://127.0.0.1:2382 ^
  --listen-client-urls http://127.0.0.1:2379 ^
  --advertise-client-urls http://127.0.0.1:2379 ^
  --initial-cluster-token etcd-cluster-1 ^
  --initial-cluster infra1=http://127.0.0.1:2381,infra2=http://127.0.0.1:2382,infra3=http://127.0.0.1:2383 ^
  --initial-cluster-state new
  
start /b cmd /c .\\etcd-v3.5.1\\etcd.exe --name infra3 --initial-advertise-peer-urls http://127.0.0.1:2383 ^
  --listen-peer-urls http://127.0.0.1:2383 ^
  --listen-client-urls http://127.0.0.1:2380 ^
  --advertise-client-urls http://127.0.0.1:2380 ^
  --initial-cluster-token etcd-cluster-1 ^
  --initial-cluster infra1=http://127.0.0.1:2381,infra2=http://127.0.0.1:2382,infra3=http://127.0.0.1:2383 ^
  --initial-cluster-state new