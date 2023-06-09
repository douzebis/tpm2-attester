#!/bin/bash
# SPDX-License-Identifier: Apache-2.0

set -e

if [ "$(uname -s)" = "Darwin" ]; then
  if [ "$(whoami)" = "root" ]; then
    TARGET_DIR="/Library/Google/Chrome/NativeMessagingHosts"
  else
    TARGET_DIR="$HOME/Library/Application Support/Google/Chrome/NativeMessagingHosts"
  fi
elif [ "$1" != "chromium" ]; then
  if [ "$(whoami)" = "root" ]; then
    TARGET_DIR="/etc/opt/chrome/native-messaging-hosts"
  else
    TARGET_DIR="$HOME/.config/google-chrome/NativeMessagingHosts"
  fi
else
  if [ "$(whoami)" = "root" ]; then
    TARGET_DIR="/etc/chromium/native-messaging-hosts"
  else
    TARGET_DIR="$HOME/.config/chromium/NativeMessagingHosts"
  fi
fi

HOST_NAME=com.google.chrome.example.echo
rm "$TARGET_DIR/com.google.chrome.example.echo.json"
echo "Native messaging host $HOST_NAME has been uninstalled."