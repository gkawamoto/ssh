FROM golang:1.14 AS builder
COPY main.go /main.go
RUN go build -o /entrypoint /main.go

FROM debian:stretch
RUN apt-get update && apt-get install -y openssh-server openssh-client
RUN mkdir /run/sshd
COPY --from=builder /entrypoint /ssh/entrypoint
ENTRYPOINT ["/ssh/entrypoint"]
