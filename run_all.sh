#!/bin/bash
# Run all fakedata generators in parallel
# Usage: ./run_all.sh [path_to_fakedata_binary]

FAKEDATA="${1:-./fakedata}"

# Array to store PIDs
pids=()

# Cleanup function
cleanup() {
    echo "Stopping all fakedata processes..."
    for pid in "${pids[@]}"; do
        kill "$pid" 2>/dev/null
    done
    wait
    echo "All processes stopped."
    exit 0
}

# Trap SIGINT and SIGTERM
trap cleanup SIGINT SIGTERM

echo "Starting fakedata generators..."

# Start each generator in background
$FAKEDATA ipfix --host 127.0.0.1 --port 4739 &
pids+=($!)
echo "Started IPFIX on port 4739 (PID: $!)"

$FAKEDATA sflow --host 127.0.0.1 --port 6343 &
pids+=($!)
echo "Started sFlow on port 6343 (PID: $!)"

$FAKEDATA syslog --host 127.0.0.1 --port 515 --type firewall &
pids+=($!)
echo "Started Syslog (firewall) on port 515 (PID: $!)"

$FAKEDATA syslog --host 127.0.0.1 --port 514 --rfc 3164 &
pids+=($!)
echo "Started Syslog (RFC 3164) on port 514 (PID: $!)"

echo ""
echo "All generators running. Press Ctrl+C to stop."
echo "PIDs: ${pids[*]}"

# Wait for all background processes
wait
