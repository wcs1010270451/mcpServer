#!/bin/bash
# restart.sh - 重启 Go 服务，不重新编译

SERVICE_NAME="McpServer"
BIN_DIR="./bin"
LOG_FILE="/tmp/${SERVICE_NAME}.log"
PID_FILE="$BIN_DIR/${SERVICE_NAME}.pid"

# 停止服务
if [ -f $PID_FILE ]; then
    PID=$(cat $PID_FILE)
    echo "Stopping $SERVICE_NAME with PID $PID..."
    kill -9 $PID
    rm -f $PID_FILE
    echo "$SERVICE_NAME stopped."
else
    echo "PID file not found. $SERVICE_NAME may not be running."
fi

# 启动服务后台运行，日志带时间戳
echo "Starting $SERVICE_NAME..."
nohup stdbuf -oL -eL $BIN_DIR/$SERVICE_NAME \
    > >(awk '{ print strftime("[%Y-%m-%d %H:%M:%S]"), $0; fflush(); }' >> $LOG_FILE) \
    2> >(awk '{ print strftime("[%Y-%m-%d %H:%M:%S] [ERR]"), $0; fflush(); }' >> $LOG_FILE) &

# 保存 PID
echo $! > $PID_FILE
echo "$SERVICE_NAME restarted with PID $(cat $PID_FILE), logs at $LOG_FILE"
