#!/bin/bash
# SPDX-License-Identifier: Apache-2.0

#!/bin/bash

POSITIONAL_ARGS=()
BROWSER=""
EXTENSION_ID=""

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

# Update host path and extension id in the manifest.
echo "Manifest for extension $HOST_NAME has been created."
if [ "$EXTENSION_ID" == "" ]; then
  if [ "$BROWSER" == "firefox" ]; then
    EXTENSION_ID=attester@example.com
    sed "s|EXTENSION_ID|$EXTENSION_ID|g" ../firefox/manifest.templ > ../firefox/manifest.json
    echo "=> ../firefox/manifest.json"
  else
    # https://stackoverflow.com/questions/23873623/obtaining-chrome-extension-id-for-development
    # (command line way)
    # Create private key called key.pem
    # 2>/dev/null openssl genrsa 2048 | openssl pkcs8 -topk8 -nocrypt -out key.pem
    # # Generate string to be used as "key" in manifest.json (outputs to stdout)
    # 2>/dev/null openssl rsa -in key.pem -pubout -outform DER | openssl base64 -A
    # # Calculate extension ID (outputs to stdout)
    # 2>/dev/null openssl rsa -in key.pem -pubout -outform DER |  shasum -a 256 | head -c32 | tr 0-9a-f a-p
    EXTENSION_ID=$(2>/dev/null openssl rsa -in key.pem -pubout -outform DER |  shasum -a 256 | head -c32 | tr 0-9a-f a-p)
    EXTENSION_KEY=$(2>/dev/null openssl rsa -in key.pem -pubout -outform DER | openssl base64 -A)
    sed "s|EXTENSION_KEY|$EXTENSION_KEY|g" ../chrome/manifest.templ > ../chrome/manifest.json
    echo "=> ../chrome/manifest.json"
  fi
fi
HOST_PATH=$DIR/attester
ESCAPED_HOST_PATH=${HOST_PATH////\\/}
cat "$DIR/$HOST_NAME.json" \
| sed "s|HOST_PATH|$ESCAPED_HOST_PATH|g" \
| sed "s|EXTENSION_ID|$EXTENSION_ID|g" \
> "$TARGET_DIR/$HOST_NAME.json"


# Set permissions for the manifest so that all users can read it.
chmod o+r "$TARGET_DIR/$HOST_NAME.json"

echo "Native messaging host $HOST_NAME has been installed."
echo "=> $TARGET_DIR/$HOST_NAME.json"