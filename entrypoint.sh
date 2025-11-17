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
        return 0
    fi
    
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
            continue
        fi
        
        # Probe peer (suppress output for periodic checks)
        gluster peer probe "$peer" > /dev/null 2>&1 || true
    done
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

# Initial peer probing with verbose output
if [ -n "$GLUSTER_PEERS" ]; then
    echo "GLUSTER_PEERS environment variable found: $GLUSTER_PEERS"
    echo "Initial peer probing..."
    probe_peers
    echo "Current peer status:"
    gluster peer status 2>&1 || echo "Could not retrieve peer status"
    echo "Peer probing will continue in background every 30 seconds"
fi

# Start the gluster-provisioner in the background and capture output
echo "Starting gluster-provisioner..."
/usr/local/bin/gluster-provisioner "$@" 2>&1 &
PROVISIONER_PID=$!
echo "Gluster-provisioner started with PID $PROVISIONER_PID"

# Monitor both processes
PROBE_COUNTER=0
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
    
    # Retry peer probing every 30 seconds (6 iterations of 5 second sleep)
    PROBE_COUNTER=$((PROBE_COUNTER + 1))
    if [ $PROBE_COUNTER -ge 6 ]; then
        probe_peers
        PROBE_COUNTER=0
    fi
    
    # Wait a bit before checking again
    sleep 5
done
