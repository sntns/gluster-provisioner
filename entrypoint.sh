#!/bin/sh
set -e

# Function to handle shutdown
shutdown() {
    echo "Shutting down..."
    if [ -n "$GLUSTERD_PID" ]; then
        kill $GLUSTERD_PID 2>/dev/null || true
    fi
    if [ -n "$PROVISIONER_PID" ]; then
        kill $PROVISIONER_PID 2>/dev/null || true
    fi
    exit 0
}

# Set up signal handlers
trap shutdown SIGTERM SIGINT

# Start GlusterFS daemon in the background and capture output
echo "Starting GlusterFS daemon..."
glusterd --no-daemon 2>&1 &
GLUSTERD_PID=$!
echo "GlusterFS daemon started with PID $GLUSTERD_PID"

# Give glusterd time to start
sleep 2

# Start the gluster-provisioner in the background and capture output
echo "Starting gluster-provisioner..."
/usr/local/bin/gluster-provisioner "$@" 2>&1 &
PROVISIONER_PID=$!
echo "Gluster-provisioner started with PID $PROVISIONER_PID"

# Monitor both processes
while true; do
    # Check if glusterd is still running
    if ! kill -0 $GLUSTERD_PID 2>/dev/null; then
        echo "ERROR: GlusterFS daemon (PID $GLUSTERD_PID) has stopped unexpectedly"
        exit 1
    fi
    
    # Check if provisioner is still running
    if ! kill -0 $PROVISIONER_PID 2>/dev/null; then
        echo "ERROR: Gluster-provisioner (PID $PROVISIONER_PID) has stopped unexpectedly"
        exit 1
    fi
    
    # Wait a bit before checking again
    sleep 5
done
