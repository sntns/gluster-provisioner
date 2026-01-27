# GlusterFS Provisioner with Multi-Daemon Docker Setup

This repository contains a Docker image that runs both the GlusterFS daemon (glusterd) and the gluster-provisioner application.

## Docker Image Versioning

The Docker image uses a dual-version tagging scheme:

- Format: `v{GLUSTER_VERSION}-{PROVISIONER_VERSION}`
- Example: `ghcr.io/sntns/gluster-provisioner:11.1-1`

Where:

- `GLUSTER_VERSION` is the version of GlusterFS cluster daemon (e.g., `11.1`)
- `PROVISIONER_VERSION` is a single-digit version of the gluster-provisioner application (e.g., `1`)

## Architecture

The Docker image is built using a multi-stage build process:

1. **Builder Stage**: Compiles the gluster-provisioner Go application
2. **GlusterFS Base Stage**: Sets up CentOS Stream 9 with GlusterFS server
3. **Final Stage**: Combines both components with an entrypoint script that:
   - Starts the GlusterFS daemon (glusterd) in the background
   - Starts the gluster-provisioner application in the background
   - Monitors both processes and merges their logs to stdout
   - Automatically terminates and exits if either process stops
   - Handles graceful shutdown of both processes
   - Includes Docker HEALTHCHECK for container orchestration

### Why CentOS Stream 9?

Alpine Linux does not support GlusterFS due to musl libc compatibility issues. CentOS Stream 9 is used as the base image because:

- Official GlusterFS packages are available
- Well-tested and supported for GlusterFS deployments
- Compatible with the Storage SIG repository

### Health Checks

The Docker image includes a built-in HEALTHCHECK that:

- Runs every 30 seconds
- Verifies both `glusterd` and `gluster-provisioner` processes are running
- Has a 10-second timeout per check
- Allows 10-second start period for initialization
- Marks container as unhealthy after 3 consecutive failures
- Enables automatic container restart policies in orchestrators (Docker Compose, Kubernetes, etc.)

## Building the Image

### Default Build

```bash
docker build -t gluster-provisioner .
```

### Custom Versions

```bash
docker build \
  --build-arg GLUSTER_VERSION=11.1 \
  --build-arg PROVISIONER_VERSION=1 \
  -t gluster-provisioner:11.1-1 \
  .
```

## CI/CD Pipeline

### Automated Builds

The repository includes two GitHub Actions workflows:

#### 1. Docker Image Build and Release (`.github/workflows/docker-image.yml`)

**On pushes to main:**

- Builds and pushes with tags:
  - `latest`
  - `{sha}`
  - `{GLUSTER_VERSION}-latest`

**On version tags (e.g., `v1`):**

- Builds and pushes with tags:
  - `{version}` (e.g., `1`)
  - `{GLUSTER_VERSION}-{version}` (e.g., `11.1-1`)
- Creates a GitHub Release with version information

**Manual Trigger:**

- Allows specifying custom GlusterFS version via workflow dispatch

#### 2. GlusterFS Version Check (`.github/workflows/check-gluster-version.yml`)

This workflow automatically checks for new GlusterFS versions:

- **Schedule**: Runs every Monday at 9:00 AM UTC
- **Process**:
  1. Checks the current GlusterFS version in the repository
  2. Queries the GlusterFS GitHub repository for the latest release
  3. If a new version is found, creates a pull request with updates to:
     - Dockerfile default version
     - CI workflow environment variable
  4. PR is automatically labeled with `dependencies` and `automated`

**Manual Trigger:**

- Can be run on-demand via workflow dispatch

## Running the Container

### Basic Run

```bash
docker run -d \
  --privileged \
  --name gluster-provisioner \
  ghcr.io/sntns/gluster-provisioner:latest \
  run
```

### With Custom Configuration

```bash
docker run -d \
  --privileged \
  --name gluster-provisioner \
  -v /path/to/config:/config \
  ghcr.io/sntns/gluster-provisioner:11.1-1 \
  run
```

### With GlusterFS Peer Configuration

```bash
docker run -d \
  --privileged \
  --name gluster-provisioner \
  -e GLUSTER_PEERS="node1.example.com,node2.example.com,192.168.1.10" \
  ghcr.io/sntns/gluster-provisioner:11.1-1 \
  run
```

The `GLUSTER_PEERS` environment variable allows you to specify a comma-separated list of GlusterFS peer addresses. When the container starts:

- The GlusterFS daemon will probe each peer in the list
- Self-references (localhost, 127.0.0.1, current hostname/IP) are automatically skipped
- Failed peers are automatically retried every 60 seconds until they connect
- Only failed peers are retried (successful connections are not re-probed)
- All probe attempts are logged for monitoring
- Peer status is displayed after initial probing completes

**Note**: The `--privileged` flag is required for the container to:

- Manage block devices
- Start the GlusterFS daemon
- Perform disk operations

### Health Check Monitoring

Check container health status:

```bash
# View health status
docker inspect --format='{{.State.Health.Status}}' gluster-provisioner

# View health check logs
docker inspect --format='{{range .State.Health.Log}}{{.Output}}{{end}}' gluster-provisioner
```

The container will automatically restart if configured with a restart policy:

```bash
docker run -d \
  --privileged \
  --restart=unless-stopped \
  --name gluster-provisioner \
  ghcr.io/sntns/gluster-provisioner:latest \
  run
```

## GlusterFS Cluster Configuration

### Peer Discovery via Environment Variables

The container supports automatic peer probing during startup using the `GLUSTER_PEERS` environment variable. This allows Docker nodes to discover and connect to each other automatically.

**Environment Variable:**

- `GLUSTER_PEERS` - Comma-separated list of peer addresses (hostnames or IPs)

**Example:**

```bash
# On node1
docker run -d \
  --privileged \
  --hostname node1 \
  --name gluster-provisioner \
  -e GLUSTER_PEERS="node2,node3,node4" \
  ghcr.io/sntns/gluster-provisioner:latest \
  run

# On node2
docker run -d \
  --privileged \
  --hostname node2 \
  --name gluster-provisioner \
  -e GLUSTER_PEERS="node1,node3,node4" \
  ghcr.io/sntns/gluster-provisioner:latest \
  run
```

**Behavior:**

- Peers are probed initially after the GlusterFS daemon starts
- Failed peers are automatically retried every 60 seconds until they connect
- Only failed peers are retried (successful connections are not re-probed)
- Self-references are automatically detected and skipped (safe to include current node)
- All probe attempts are logged for monitoring
- Peer status is displayed after initial probing completes
- Works with hostnames, IP addresses, or FQDNs

**Docker Compose Example:**

```yaml
version: '3.8'
services:
  gluster-node1:
    image: ghcr.io/sntns/gluster-provisioner:latest
    privileged: true
    hostname: node1
    environment:
      - GLUSTER_PEERS=node2,node3
    command: run

  gluster-node2:
    image: ghcr.io/sntns/gluster-provisioner:latest
    privileged: true
    hostname: node2
    environment:
      - GLUSTER_PEERS=node1,node3
    command: run

  gluster-node3:
    image: ghcr.io/sntns/gluster-provisioner:latest
    privileged: true
    hostname: node3
    environment:
      - GLUSTER_PEERS=node1,node2
    command: run
```

## Available Commands

The gluster-provisioner supports several commands:

- `run` - Run the gluster-provisioner listener (default daemon mode)
- `enumerate` - Enumerate all devices on the system
- `partition` - Partition a disk
- `format` - Format a disk
- `mount` - Mount a filesystem
- `umount` - Unmount a filesystem

## Version Information

To check the versions of components in the image:

```bash
# Check image labels
docker inspect ghcr.io/sntns/gluster-provisioner:11.1-1 | jq '.[0].Config.Labels'

# Example output:
{
  "gluster.version": "11.1",
  "provisioner.version": "1",
  "org.opencontainers.image.title": "GlusterFS Provisioner",
  "org.opencontainers.image.description": "Container running GlusterFS daemon and provisioner"
}
```

## Updating GlusterFS Version

### Automatic Updates

The automated workflow checks for updates weekly and creates PRs.

### Manual Updates

1. Update `GLUSTER_VERSION` in `Dockerfile`
2. Update `GLUSTER_VERSION` in `.github/workflows/docker-image.yml`
3. Commit and push changes
4. Tag a new release: `git tag -a v2 -m "Provisioner version 2"`
5. Push the tag: `git push origin v2`

## Development

### Prerequisites

- Go 1.24 or later
- Docker 20.10 or later
- Access to CentOS/GlusterFS repositories (for building)

### Local Development

```bash
# Build the Go application locally
cd cmd/gluster-provisioner
go build -o gluster-provisioner

# Run locally (requires glusterd to be running)
./gluster-provisioner run
```

### Testing

```bash
# Run Go tests
go test ./...

# Build test image
docker build -t test-gluster-provisioner .
```

## Architecture Decisions

### Multi-Process Container

While generally discouraged, this container runs multiple processes (glusterd + provisioner) because:

1. The provisioner requires a running GlusterFS daemon to function
2. They are tightly coupled and must run on the same host
3. The entrypoint script manages both processes appropriately

### Base Image Choice

CentOS Stream 9 was chosen over Alpine Linux because:

- GlusterFS has no official Alpine packages
- musl libc incompatibilities prevent building GlusterFS on Alpine
- CentOS Stream has official GlusterFS support from the Storage SIG

## Troubleshooting

### GlusterFS daemon not starting

```bash
# Check logs
docker logs gluster-provisioner

# Ensure container has privileges
docker run --privileged ...
```

### Provisioner can't connect to glusterd

The entrypoint script includes a 2-second delay for glusterd startup. If issues persist, you may need to adjust this in `entrypoint.sh`.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

For security-related reports, see [SECURITY.md](SECURITY.md).

## Maintainer / Contact

This repository is maintained by **Sentiens**. For inquiries, contact: `contact@sentiens.fr`.
