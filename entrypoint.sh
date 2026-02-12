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

# Global variable to track failed peers
FAILED_PEERS=""

# Function to probe gluster peers (initial probe)
probe_peers_initial() {
    if [ -z "$GLUSTER_PEERS" ]; then
        return 0
    fi
    
    # Get current hostname/IP to avoid self-probing issues
    CURRENT_HOSTNAME=$(hostname)
    CURRENT_IP=$(hostname -i 2>/dev/null || echo "")
    
    FAILED_PEERS=""
    
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
            echo "Skipping self-peer: $peer"
            continue
        fi
        
        # Probe peer
        echo "Probing peer: $peer"
        if gluster peer probe "$peer" 2>&1; then
            echo "Successfully probed peer: $peer"
        else
            echo "Failed to probe peer: $peer"
            # Add to failed peers list (in parent shell scope via file)
            echo "$peer" >> /tmp/failed_peers.txt
        fi
    done
    
    # Read failed peers from temp file
    if [ -f /tmp/failed_peers.txt ]; then
        FAILED_PEERS=$(tr '\n' ',' < /tmp/failed_peers.txt | sed 's/,$//')
        rm -f /tmp/failed_peers.txt
    fi
}

# Function to retry failed peers only
retry_failed_peers() {
    if [ -z "$FAILED_PEERS" ]; then
        return 0
    fi
    
    # Get current hostname/IP to avoid self-probing issues
    CURRENT_HOSTNAME=$(hostname)
    CURRENT_IP=$(hostname -i 2>/dev/null || echo "")
    
    # Split FAILED_PEERS by comma and retry each
    echo "$FAILED_PEERS" | tr ',' '\n' | while read -r peer; do
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
        
        # Retry probing peer
        echo "Retrying failed peer: $peer"
        if gluster peer probe "$peer" > /dev/null 2>&1; then
            echo "Successfully probed peer: $peer"
        else
            echo "Still failed to probe peer: $peer"
            # Keep in failed peers list
            echo "$peer" >> /tmp/failed_peers_new.txt
        fi
    done
    
    # Update failed peers list
    if [ -f /tmp/failed_peers_new.txt ]; then
        FAILED_PEERS=$(tr '\n' ',' < /tmp/failed_peers_new.txt | sed 's/,$//')
        rm -f /tmp/failed_peers_new.txt
    else
        FAILED_PEERS=""
    fi
}

# Set up signal handlers
trap shutdown SIGTERM SIGINT

# Start GlusterFS daemon in the background and capture output
echo "Starting GlusterFS daemon..."
glusterd --no-daemon 2>&1 &
GLUSTERD_PID=$!
echo "GlusterFS daemon started with PID $GLUSTERD_PID"

# Give glusterd time to start
sleep 5

# Initial peer probing with verbose output
if [ -n "$GLUSTER_PEERS" ]; then
    echo "GLUSTER_PEERS environment variable found: $GLUSTER_PEERS"
    echo "Initial peer probing..."
    probe_peers_initial
    echo "Current peer status:"
    gluster peer status 2>&1 || echo "Could not retrieve peer status"
    if [ -n "$FAILED_PEERS" ]; then
        echo "Failed peers will be retried every 60 seconds: $FAILED_PEERS"
    else
        echo "All peers successfully probed"
    fi
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
    
    # Retry failed peers every 60 seconds (12 iterations of 5 second sleep)
    PROBE_COUNTER=$((PROBE_COUNTER + 1))
    if [ $PROBE_COUNTER -ge 12 ] && [ -n "$FAILED_PEERS" ]; then
        retry_failed_peers
        PROBE_COUNTER=0
    fi
    
    # Wait a bit before checking again
    sleep 5
done
