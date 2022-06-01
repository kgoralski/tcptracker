FROM golang:alpine as builder
ENV GO111MODULE=on

ARG APP_HOME=/go/src/tcptracker
ADD . "$APP_HOME"
WORKDIR "$APP_HOME"

RUN apk add gcc libc-dev libpcap-dev

RUN go mod download
RUN go mod verify
RUN go build -o tcptracker cmd/main.go

FROM golang:alpine
ENV GO111MODULE=on

RUN apk update && \
    apk add --no-cache iptables ip6tables gcc libc-dev libpcap-dev dpkg

# Quick hack to use iptables-nft using dpkg, that shouldn't be inside container
RUN update-alternatives --install /sbin/iptables iptables /sbin/iptables-nft 10

ENV APP_HOME=/go/src/tcptracker
RUN mkdir -p "$APP_HOME"

WORKDIR "$APP_HOME"

COPY --from=builder "$APP_HOME"/tcptracker $APP_HOME

EXPOSE 8081
ENTRYPOINT ["./tcptracker"]