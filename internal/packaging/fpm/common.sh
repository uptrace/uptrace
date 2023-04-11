#!/bin/shell

PKG_NAME="uptrace"
PKG_VENDOR="Uptrace"
PKG_MAINTAINER="Vladimir Mihailenco <vladimir.webdev@gmail.com>"
PKG_DESCRIPTION="Open Source Observability with Traces, Metrics, and Logs"
PKG_LICENSE="BSL"
PKG_URL="https://github.com/uptrace/uptrace"
PKG_USER="uptrace"
PKG_GROUP="uptrace"

SERVICE_NAME="uptrace"
PROCESS_NAME="uptrace"
FPM_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"

cp $REPO_DIR/config/uptrace.yml $FPM_DIR/uptrace.yml

CONFIG_PATH="$FPM_DIR/uptrace.yml"
SERVICE_PATH="$FPM_DIR/$SERVICE_NAME.service"
ENVFILE_PATH="$FPM_DIR/$SERVICE_NAME.conf"
PREINSTALL_PATH="$FPM_DIR/preinstall.sh"
POSTINSTALL_PATH="$FPM_DIR/postinstall.sh"
PREUNINSTALL_PATH="$FPM_DIR/preuninstall.sh"
