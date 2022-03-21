#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
REPO_DIR="$( cd "$SCRIPT_DIR/../../../../" && pwd )"
VERSION="${1:-}"
ARCH="${2:-"amd64"}"
OUTPUT_DIR="${3:-"$REPO_DIR/dist/"}"

source $SCRIPT_DIR/../common.sh
UPTRACEBIN_PATH="$REPO_DIR/bin/${SERVICE_NAME}_linux_$ARCH"

# remap arm64 to aarch64, which is the arch used by Linux distributions
if [[ "$ARCH" == "arm64" ]]; then
    ARCH="aarch64"
fi

mkdir -p "$OUTPUT_DIR"

fpm -s dir -t rpm -n $PKG_NAME -v ${VERSION#v} -f -p "$OUTPUT_DIR" \
    --vendor "$PKG_VENDOR" \
    --maintainer "$PKG_MAINTAINER" \
    --description "$PKG_DESCRIPTION" \
    --license "$PKG_LICENSE" \
    --url "$PKG_URL" \
    --architecture "$ARCH" \
    --deb-dist "stable" \
    --deb-user "$PKG_USER" \
    --deb-group "$PKG_GROUP" \
    --before-install "$PREINSTALL_PATH" \
    --after-install "$POSTINSTALL_PATH" \
    --pre-uninstall "$PREUNINSTALL_PATH" \
    $SERVICE_PATH=/lib/systemd/system/$SERVICE_NAME.service \
    $UPTRACEBIN_PATH=/usr/bin/$SERVICE_NAME \
    $CONFIG_PATH=/etc/$SERVICE_NAME/uptrace.yml \
    $ENVFILE_PATH=/etc/$SERVICE_NAME/$SERVICE_NAME.conf
