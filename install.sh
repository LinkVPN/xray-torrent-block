#!/bin/bash

set -e

if [ "$EUID" -ne 0 ]; then
  echo "Please run the script with root privileges (sudo)."
  exit 1
fi

echo "Installing necessary dependencies..."
if command -v apt-get &> /dev/null; then
  apt-get update -qq
  apt-get install -y ufw curl > /dev/null
elif command -v yum &> /dev/null; then
  yum install -y epel-release > /dev/null
  yum install -y ufw curl > /dev/null
elif command -v dnf &> /dev/null; then
  dnf install -y ufw curl > /dev/null
elif command -v pacman &> /dev/null; then
  pacman -Sy --noconfirm ufw curl > /dev/null
else
  echo "Iptables will be used. Please install ufw and curl manually."
fi

if systemctl is-active --quiet torrent-blocker; then
  echo "Stopping existing torrent-blocker service..."
  systemctl stop torrent-blocker
fi

ARCH=""
if [ "$(uname -m)" == "x86_64" ]; then
  ARCH="amd64"
elif [ "$(uname -m)" == "aarch64" ];then
  ARCH="arm64"
else
  echo "Unsupported architecture."
  exit 1
fi

echo "Downloading the latest version of torrent-blocker..."
LATEST_RELEASE=$(curl -sL https://api.github.com/repos/LinkVPN/xray-torrent-block/releases/latest | grep tag_name | cut -d '"' -f 4)
URL="https://github.com/LinkVPN/xray-torrent-block/releases/download/${LATEST_RELEASE}/xray-torrent-block-${LATEST_RELEASE}-linux-${ARCH}.tar.gz"

curl -sL "$URL" -o /opt/torrent-blocker.tar.gz

echo "Extracting files..."
mkdir -p /opt/torrent-blocker
tar -xzf /opt/torrent-blocker.tar.gz -C /opt/torrent-blocker --overwrite
rm /opt/torrent-blocker.tar.gz

CONFIG_PATH="/opt/torrent-blocker/config.yaml"
CONFIG_TEMPLATE_PATH="/opt/torrent-blocker/config.yaml.example"

if [ ! -f "$CONFIG_PATH" ]; then
  mv "$CONFIG_TEMPLATE_PATH" "$CONFIG_PATH"
  echo "New configuration file created at $CONFIG_PATH"
else
  echo "Configuration file already exists. Checking its contents..."
fi

check_placeholder() {
  local key="$1"
  grep -q "$key: \"ADMIN_" "$CONFIG_PATH"
}

ask_for_input=false

echo "Setting up systemd service..."
curl -sL https://github.com/LinkVPN/xray-torrent-block/raw/refs/heads/main/torrent-blocker.service -o /etc/systemd/system/torrent-blocker.service

systemctl daemon-reload
systemctl enable torrent-blocker
systemctl start torrent-blocker

echo ""
echo "==============================================================="
echo ""
echo "Installation complete! The torrent-blocker service is running."
echo ""
echo "==============================================================="
echo ""
echo "You can configure additional options in the configuration file"
echo "/opt/torrent-blocker/config.yaml"
echo "It is possible to enable sending user notifications via Telegram."
echo ""
echo "==============================================================="
