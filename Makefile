.PHONY: build run gorun compile lint test test_race test_coverage check_coverage mocks dep vet errcheck install_tools docker_build docker_run docker_run_detach docker_start docker_remove docker_exec

BINARY_NAME=tcptracker
DEVICE=eth0

build:
	go build -o bin/${BINARY_NAME} ./cmd/main.go

gorun:
	go run ./cmd/main.go

run:
	sudo setcap cap_net_admin,cap_net_raw+ep ./bin/tcptracker
	sudo ./bin/tcptracker -deviceName ${DEVICE}

fmt:
	go fmt ./...

lint:
	golangci-lint --version
	golangci-lint --config ./.golangci.yml linters
	golangci-lint --config ./.golangci.yml run --timeout 5m

test:
	go test -v ./...

test_race:
	go test -v -race ./...

test_coverage:
	go test -v ./... -coverprofile=coverage.out

check_coverage:
	go tool cover -html=coverage.out

mocks: # TODO make it dynamic
	mockgen -source=internal/connectiontracker/firewall.go -package=mock -destination=mock/gomock_firewall.go Firewall
	mockgen -source=internal/connectiontracker/firewall_test.go -package=mock -destination=mock/gomock_ipTableCoreos.go ipTableCoreos

dep:
	go mod tidy

vet:
	go vet ./...

errcheck:
	errcheck ./...

install_tools:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.46.2
	golangci-lint --version
	go install github.com/kisielk/errcheck@latest
	go install github.com/golang/mock/mockgen@v1.6.0

docker_build:
	docker build -t ${BINARY_NAME} .

docker_run:
	docker run --name ${BINARY_NAME} --net=host --cap-add NET_ADMIN -t tcptracker:latest -deviceName ${DEVICE}

docker_run_detach:
	docker run -d --name ${BINARY_NAME} --net=host --cap-add NET_ADMIN -t tcptracker:latest -deviceName ${DEVICE}

docker_start:
	docker start tcptracker

docker_exec:
	docker exec -it tcptracker sh

docker_remove:
	docker rm tcptracker