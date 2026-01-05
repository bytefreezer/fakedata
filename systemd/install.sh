#!/bin/bash
# Install ByteFreezer Fakedata systemd services
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_PATH="${1:-/usr/local/bin/bytefreezer-fakedata}"

echo "Installing ByteFreezer Fakedata services..."

# Check if binary exists
if [ ! -f "$BINARY_PATH" ]; then
    echo "Binary not found at $BINARY_PATH"
    echo "Building from source..."
    cd "$SCRIPT_DIR/.."
    go build -o bytefreezer-fakedata .
    sudo cp bytefreezer-fakedata /usr/local/bin/
    echo "Binary installed to /usr/local/bin/bytefreezer-fakedata"
fi

# Copy service files
sudo cp "$SCRIPT_DIR"/*.service /etc/systemd/system/
sudo cp "$SCRIPT_DIR"/*.target /etc/systemd/system/

# Reload systemd
sudo systemctl daemon-reload

# Enable services
sudo systemctl enable bytefreezer-fakedata.target
sudo systemctl enable bytefreezer-fakedata-ipfix.service
sudo systemctl enable bytefreezer-fakedata-sflow.service
sudo systemctl enable bytefreezer-fakedata-syslog-firewall.service
sudo systemctl enable bytefreezer-fakedata-syslog-rfc3164.service

echo ""
echo "Installation complete!"
echo ""
echo "Commands:"
echo "  Start all:    sudo systemctl start bytefreezer-fakedata.target"
echo "  Stop all:     sudo systemctl stop bytefreezer-fakedata.target"
echo "  Status:       sudo systemctl status 'bytefreezer-fakedata-*'"
echo "  Logs:         journalctl -u 'bytefreezer-fakedata-*' -f"
echo ""
echo "Individual services:"
echo "  sudo systemctl start bytefreezer-fakedata-ipfix"
echo "  sudo systemctl start bytefreezer-fakedata-sflow"
echo "  sudo systemctl start bytefreezer-fakedata-syslog-firewall"
echo "  sudo systemctl start bytefreezer-fakedata-syslog-rfc3164"
