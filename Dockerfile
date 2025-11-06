FROM golang:1.21-alpine as builder

WORKDIR /app

COPY . .

RUN cd cmd/gluster-provisioner && go build -o /gluster-provisioner

FROM alpine:3.19

RUN apk add --no-cache eudev parted e2fsprogs lsblk

COPY --from=builder /gluster-provisioner /usr/local/bin/gluster-provisioner

ENTRYPOINT ["/usr/local/bin/gluster-provisioner"]