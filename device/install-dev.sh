#!/bin/bash
# SPDX-License-Identifier: Apache-2.0

#!/bin/bash

POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--help)
      EXTENSION_ID="$2"
      shift # past argument
      shift # past value
      cat <<'EOF'
Usage: install-dev.sh [OPTION]...
Install native-messaging manifest file on host

  -b, --browser            chrome | chromium | firefox
  -i, --extension-id       extension id (e.g. ejjoloepomkaefacigjcjpedphnlflpn)
EOF
      ;;
    -b|--browser)
      BROWSER="$2"
      shift # past argument
      shift # past value
      ;;
    -i|--extension-id)
      EXTENSION_ID="$2"
      shift # past argument
      shift # past value
      ;;
    -*|--*)
      echo "Unknown option $1"
      exit 1
      ;;
    *)
      POSITIONAL_ARGS+=("$1") # save positional arg
      shift # past argument
      ;;
  esac
done

set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters

set -e

DIR="$( cd "$( dirname "$0" )" && pwd )"
if [ "$(uname -s)" = "Darwin" ]; then
  if [ "$(whoami)" = "root" ]; then
    TARGET_DIR="/Library/Google/Chrome/NativeMessagingHosts"
  else
    TARGET_DIR="$HOME/Library/Application Support/Google/Chrome/NativeMessagingHosts"
  fi
elif [ "$BROWSER" == "chromium" ]; then
  if [ "$(whoami)" = "root" ]; then
    TARGET_DIR="/etc/chromium/native-messaging-hosts"
  else
    TARGET_DIR="$HOME/.config/chromium/NativeMessagingHosts"
  fi
elif [ "$BROWSER" == "chrome" ]; then
  if [ "$(whoami)" = "root" ]; then
    TARGET_DIR="/etc/opt/chrome/native-messaging-hosts"
  else
    TARGET_DIR="$HOME/.config/google-chrome/NativeMessagingHosts"
  fi
elif [ "$BROWSER" == "firefox" ]; then
  if [ "$(whoami)" = "root" ]; then
    TARGET_DIR="/usr/lib/mozilla/native-messaging-hosts"
  else
    TARGET_DIR="$HOME/.mozilla/native-messaging-hosts"
  fi
else
  echo "Unknown browser type: $BROWSER" >&2
  exit 1
fi

HOST_NAME=com.douzebis.attester

# Create directory to store native messaging host.
mkdir -p "$TARGET_DIR"

# Copy native messaging host manifest.
# cp "$DIR/$HOST_NAME.json" "$TARGET_DIR"

# Update host path in the manifest.
HOST_PATH=$DIR/attester
ESCAPED_HOST_PATH=${HOST_PATH////\\/}
cat "$DIR/$HOST_NAME.json" \
| sed "s/HOST_PATH/$ESCAPED_HOST_PATH/" \
| sed "s/EXTENSION_ID/$EXTENSION_ID/" \
> "$TARGET_DIR/$HOST_NAME.json"

# Set permissions for the manifest so that all users can read it.
chmod o+r "$TARGET_DIR/$HOST_NAME.json"

echo "Native messaging host $HOST_NAME has been installed."
echo "=> $TARGET_DIR/$HOST_NAME.json"