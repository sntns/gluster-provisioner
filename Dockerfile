# Build arguments for versioning
ARG GLUSTER_VERSION=11.2
ARG PROVISIONER_VERSION=1

FROM golang:1.24-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /app

COPY . .

RUN cd cmd/gluster-provisioner && go build -o /gluster-provisioner

# Use CentOS as base for GlusterFS support (Alpine doesn't support GlusterFS)
FROM quay.io/centos/centos:stream9 AS gluster-base

ARG GLUSTER_VERSION

# Install GlusterFS from CentOS Storage SIG repository
RUN dnf install -y centos-release-gluster${GLUSTER_VERSION%%.*} && \
    dnf install -y glusterfs-server && \
    dnf clean all

# Final stage
FROM gluster-base

ARG GLUSTER_VERSION
ARG PROVISIONER_VERSION

WORKDIR /app

# Install additional tools needed by provisioner
RUN dnf install -y \
    parted \
    e2fsprogs \
    util-linux \
    hostname \
    udev \
    samba \
    which && \
    dnf clean all

# Copy the provisioner binary from builder
COPY --from=builder /gluster-provisioner /usr/local/bin/gluster-provisioner

# Copy default configuration files (expected at ./config by the binary)
COPY config/ /app/config/

# Copy the entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Copy the healthcheck script
COPY healthcheck.sh /healthcheck.sh
RUN chmod +x /healthcheck.sh

# Add version labels
LABEL org.opencontainers.image.title="GlusterFS Provisioner"
LABEL org.opencontainers.image.description="Container running GlusterFS daemon and provisioner"
LABEL gluster.version="${GLUSTER_VERSION}"
LABEL provisioner.version="${PROVISIONER_VERSION}"

# Configure health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD /healthcheck.sh

ENTRYPOINT ["/entrypoint.sh"]