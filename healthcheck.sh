#!/bin/sh
# Health check script for Docker container
# Returns 0 if both processes are running, 1 otherwise

# Check if glusterd is running
if ! pgrep -x glusterd > /dev/null; then
    echo "HEALTHCHECK FAILED: glusterd is not running"
    exit 1
fi

# Check if gluster-provisioner is running
if ! pgrep -f gluster-provisioner > /dev/null; then
    echo "HEALTHCHECK FAILED: gluster-provisioner is not running"
    exit 1
fi

# Both processes are running
exit 0
