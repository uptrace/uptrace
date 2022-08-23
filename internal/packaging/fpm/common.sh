#!/bin/shell

PKG_NAME="uptrace"
PKG_VENDOR="Uptrace"
PKG_MAINTAINER="Vladimir Mihailenco <vladimir.webdev@gmail.com>"
PKG_DESCRIPTION="Distributed tracing, logs, and metrics using OpenTelemetry and ClickHouse"
PKG_LICENSE="BSL"
PKG_URL="https://github.com/uptrace/uptrace"
PKG_USER="uptrace"
PKG_GROUP="uptrace"

SERVICE_NAME="uptrace"
PROCESS_NAME="uptrace"

FPM_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
CONFIG_PATH="$REPO_DIR/config/uptrace.yml"
SERVICE_PATH="$FPM_DIR/$SERVICE_NAME.service"
ENVFILE_PATH="$FPM_DIR/$SERVICE_NAME.conf"
PREINSTALL_PATH="$FPM_DIR/preinstall.sh"
POSTINSTALL_PATH="$FPM_DIR/postinstall.sh"
PREUNINSTALL_PATH="$FPM_DIR/preuninstall.sh"
