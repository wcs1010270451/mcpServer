#!/bin/bash
# stop.sh - 关闭 Go 服务

SERVICE_NAME="McpServer"
BIN_DIR="./bin"
PID_FILE="$BIN_DIR/${SERVICE_NAME}.pid"

if [ ! -f $PID_FILE ]; then
    echo "PID file not found. Is $SERVICE_NAME running?"
    exit 1
fi

PID=$(cat $PID_FILE)
echo "Stopping $SERVICE_NAME with PID $PID..."
kill -9 $PID

# 删除 PID 文件
rm -f $PID_FILE
echo "$SERVICE_NAME stopped."
