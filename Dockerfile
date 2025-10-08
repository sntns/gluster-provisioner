FROM golang:1.21-alpine as builder

WORKDIR /app

COPY . .

RUN cd cmd/gluster-provisioner-listener && go build -o /gluster-provisioner-listener

FROM alpine:3.19

RUN apk add --no-cache eudev parted e2fsprogs

COPY --from=builder /gluster-provisioner-listener /usr/local/bin/gluster-provisioner-listener

ENTRYPOINT ["/usr/local/bin/gluster-provisioner-listener"]