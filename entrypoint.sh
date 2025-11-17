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

# Function to probe gluster peers
probe_peers() {
    if [ -z "$GLUSTER_PEERS" ]; then
        echo "No GLUSTER_PEERS environment variable set, skipping peer probing"
        return 0
    fi
    
    echo "GLUSTER_PEERS environment variable found: $GLUSTER_PEERS"
    
    # Get current hostname/IP to avoid self-probing issues
    CURRENT_HOSTNAME=$(hostname)
    CURRENT_IP=$(hostname -i 2>/dev/null || echo "")
    
    # Split GLUSTER_PEERS by comma and probe each peer
    echo "$GLUSTER_PEERS" | tr ',' '\n' | while read -r peer; do
        # Trim whitespace
        peer=$(echo "$peer" | xargs)
        
        # Skip empty entries
        if [ -z "$peer" ]; then
            continue
        fi
        
        # Check if peer is the current node (by hostname or IP)
        if [ "$peer" = "$CURRENT_HOSTNAME" ] || [ "$peer" = "$CURRENT_IP" ] || [ "$peer" = "localhost" ] || [ "$peer" = "127.0.0.1" ]; then
            echo "Skipping self-peer: $peer (current node)"
            continue
        fi
        
        echo "Probing peer: $peer"
        if gluster peer probe "$peer" 2>&1; then
            echo "Successfully probed peer: $peer"
        else
            # GlusterFS may return errors for peers that are already probed or self-references
            # We log but don't fail the container startup
            echo "Warning: Failed to probe peer $peer (may already be connected or unreachable)"
        fi
    done
    
    # Show peer status
    echo "Current peer status:"
    gluster peer status 2>&1 || echo "Could not retrieve peer status"
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

# Wait a bit more for glusterd to be fully ready
sleep 3

# Probe gluster peers if configured
probe_peers

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
        # Kill the other process
        kill $PROVISIONER_PID 2>/dev/null || true
        exit 1
    fi
    
    # Check if provisioner is still running
    if ! kill -0 $PROVISIONER_PID 2>/dev/null; then
        echo "ERROR: Gluster-provisioner (PID $PROVISIONER_PID) has stopped unexpectedly"
        # Kill the other process
        kill $GLUSTERD_PID 2>/dev/null || true
        exit 1
    fi
    
    # Wait a bit before checking again
    sleep 5
done
