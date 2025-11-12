#!/bin/sh
set -e

# Start GlusterFS daemon in the background
echo "Starting GlusterFS daemon..."
glusterd

# Give glusterd time to start
sleep 2

# Run the gluster-provisioner
echo "Starting gluster-provisioner..."
exec /usr/local/bin/gluster-provisioner "$@"
