# TCP Tracker

* New Host connections tracking from interface like `eth0` using `google/gopacket (pcap)` library
  * Logging new connections to console output
  * Functionality is limited to read the ipv4 layer
* HTTP server with `/metrics` endpoint and new connections counter `tcptracker_new_connections`
  * Using locally, `8081` port, `http://localhost:8081/metrics`
* Using BPF Filter `tcp[tcpflags] &(tcp-syn) != 0 and tcp[tcpflags] &(tcp-ack) = 0`
* Port scan detection
  * Single source IP connects to more than 3 host ports in the previous minute
  * Blocking source ip using Firewall/IPtables using separate chain `tcptracker`
* Using fast cache with 1 minute TTL to expire connections 
* Application tries to get `Host IP Address` on start up to put it on `allow list`
(because we are checking inbound and outbound traffic)


## Tech stack

* golang 1.8 
* go modules for dependency management
* go-chi - lightweight, idiomatic and composable router for building Go HTTP services
* google/gopacket - Provides packet processing capabilities for Go (pcap) 
* coreos/go-iptables - library to manage iptables
* eko/gocache - cache manager
* dgraph-io/ristretto - cache implementation, a high performance memory-bound Go cache
  * It provides the TTL functionality
  * It's thread safe
* rs/zerolog - logging with minimum allocations
* prometheus/client_golang - metrics
* testing - testify assertions, google/gomock mocks, go-cmp - easy comparisons
* Makefile - for easy build / test scripting, most of the development is done locally, it is good enough to get fast feedback
* Dockerfile - for containerization
* golangci-lint linters

To download Golang dependencies `make dep`
```makefile
dep:
	go mod tidy
```

## How To's

### Dependencies

Libraries below are needed to run or build the application, `libpcap` and `iptables`
`iptables` is usually pre-installed. Package on my OS is `iptables-nft`. 
Packages on the Host and in docker container should match, if not some tables might be visible with different binary when checking `iptables -vnL`

apt-get dnf package manager
```
apt-get install libpcap libpcap-dev iptables
```

dnf package manager
```
dnf install libpcap libpcap-devel iptables
```

You can check the installed `iptables` package with

```
sudo update-alternatives --config iptables

There is 1 program that provides 'iptables'.

  Selection    Command
-----------------------------------------------
*+ 1           /usr/sbin/iptables-nft

```

### Building and Running locally

#### How to build a binary 

`make build`
```makefile
BINARY_NAME=tcptracker

build:
	go build -o bin/${BINARY_NAME} ./cmd/main.go
```

#### How to run a binary

Running without sudo https://dbpilot.net/3-ways-to-list-all-iptables-rules-by-a-non-root-user/

`make gorun` - to run with Go
`make run` - to run a binary from `./bin/tcptracker`

To pass the device interface name to binary you can use `-devicename eth0` flag
```make DEVICE=eth0 run```

```makefile
DEVICE=eth0

gorun:
	sudo setcap cap_net_admin,cap_net_raw+ep go run ./cmd/main.go

run:
	sudo setcap cap_net_admin,cap_net_raw+ep ./bin/tcptracker
	sudo ./bin/tcptracker -deviceName ${DEVICE}
```

#### Docker

Docker needs to be installed on the machine and docker daemon needs to be running

#### How to build a docker image

`make docker_build`

```makefile
BINARY_NAME=tcptracker

docker_build:
	docker build -t ${BINARY_NAME} .
```
#### How to run a docker container

`make DEVICE=eth0 docker_run`

```makefile
BINARY_NAME=tcptracker
DEVICE=eth0

docker_run:
	docker run --name ${BINARY_NAME} --net=host --cap-add NET_ADMIN -t tcptracker:latest -deviceName ${DEVICE}

docker_run_detach:
	docker run -d --name ${BINARY_NAME} --net=host --cap-add NET_ADMIN -t tcptracker:latest -deviceName ${DEVICE}

docker_start:
	docker start tcptracker

docker_exec:
	docker exec -it tcptracker sh
```

## Additional permissions for awareness

* Running a binary requires additional permissions `setcap cap_net_admin,cap_net_raw+ep ${BINARY_NAME}`
* Running a docker with a privileges to Host network requires NET_ADMIN `docker run --name ${BINARY_NAME} --net=host --cap-add NET_ADMIN tcptracker:latest`

## Shortcuts taken
* Added `TODOs` in source code to document _shortcuts_
* More tests in table tests style
* Concurrent tests are quite basic with a room for improvements
* Code structure could be probably more modular, but depends on the new requirements I would refactor it
* Dockerfile needs some work to have more consistent behaviour with `iptables` packages
* Running docker container as root - would need some work to run it in rootless mode
* `RUN update-alternatives --install /sbin/iptables iptables /sbin/iptables-nft 10` inside the Dockerfile, my OS is using this by default
* I would add better Configuration of the application with env variables, in 12factor manner
* I would add CI like Github Actions
* Not considered as production ready, something to take look https://github.com/kgoralski/microservice-production-readiness-checklist (but that's for backend development)

## Known issues

1. permissions to filter table 
> "running [/sbin/iptables -t filter -S tcptracker 1 --wait]: exit status 3: iptables v1.8.8 (legacy): can't initialize iptables table `filter': Permission denied\nPerhaps iptables or your kernel needs to be upgraded.\n"

fix `sudo modprobe iptable_filter`

Found it somewhere here https://github.com/nuBacuk/docker-openvpn-arm64#errors


2. iptables device is busy / linked etc.

```
$ sudo iptables -L INPUT --line-numbers -vn
$ sudo iptables -D INPUT <line_number> # li
```
or even remove whole chain
```
$ sudo iptables --flush tcptrackerNew
$ sudo iptables -X tcptrackerNew
```