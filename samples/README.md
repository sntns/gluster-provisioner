```markdown
## GlusterFS Cluster with Docker Compose

See `docker-compose.yml` for a complete example of a 3-node GlusterFS cluster using the `GLUSTER_PEERS` environment variable for automatic peer discovery.

To start the cluster:

```bash
docker-compose up -d
```

To verify the cluster is formed:

```bash
# Check logs to see peer probing
docker-compose logs gluster-node1

# Exec into a node to check peer status
docker exec gluster-node1 gluster peer status
```

The `GLUSTER_PEERS` environment variable accepts:
- Comma-separated list of hostnames or IP addresses
- Can safely include the current node (will be automatically skipped)
- Self-references like `localhost` or `127.0.0.1` are automatically detected and skipped

## Allocate a 4GB file

```bash
fallocate -l 4G /tmp/gluster-block.img
```

## Attach the file as a loop device

```bash
LOOPDEV=$(losetup --show -f /tmp/gluster-block.img)
echo "Loop device: $LOOPDEV"
```

## Generate `device.json` with the block device

```bash
cat <<EOF > device.json
{
    "Name": "gluster-block-device",
    "Path": "$LOOPDEV"
}
EOF
```

## Detach the loop device when done

```bash
losetup -d $LOOPDEV
```