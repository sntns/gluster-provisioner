```markdown
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