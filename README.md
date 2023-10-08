goserver is a micro service framework for game backend,  implemented by golang.

# Features
- network implemented by std lib on windows, and by epoll on linux to reduce goroutines scheduling.
- registry interface for micro service.
- basic rpc infrastructure, with optional load balance routers.
- data serialized through protobuf, auto generated rpc message mappings by tool.
- local actor frame to avoid mutex lock.
- utilities such as logger , fsm, timer, ring, queue, bucketmap.

# TODO
- k8s helm deployment.
- remote virtual actor enveloped by rpc.

