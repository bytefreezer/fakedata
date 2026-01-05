#!/bin/bash
# Uninstall ByteFreezer Fakedata systemd services
set -e

echo "Uninstalling ByteFreezer Fakedata services..."

# Stop services
sudo systemctl stop bytefreezer-fakedata.target 2>/dev/null || true
sudo systemctl stop bytefreezer-fakedata-ipfix.service 2>/dev/null || true
sudo systemctl stop bytefreezer-fakedata-sflow.service 2>/dev/null || true
sudo systemctl stop bytefreezer-fakedata-syslog-firewall.service 2>/dev/null || true
sudo systemctl stop bytefreezer-fakedata-syslog-rfc3164.service 2>/dev/null || true

# Disable services
sudo systemctl disable bytefreezer-fakedata.target 2>/dev/null || true
sudo systemctl disable bytefreezer-fakedata-ipfix.service 2>/dev/null || true
sudo systemctl disable bytefreezer-fakedata-sflow.service 2>/dev/null || true
sudo systemctl disable bytefreezer-fakedata-syslog-firewall.service 2>/dev/null || true
sudo systemctl disable bytefreezer-fakedata-syslog-rfc3164.service 2>/dev/null || true

# Remove service files
sudo rm -f /etc/systemd/system/bytefreezer-fakedata*.service
sudo rm -f /etc/systemd/system/bytefreezer-fakedata*.target

# Reload systemd
sudo systemctl daemon-reload

echo "Uninstallation complete!"
