#!/bin/bash
# start.sh - 启动 Go 服务并记录日志

SERVICE_NAME="McpServer"
BIN_DIR="./bin"
LOG_FILE="/tmp/${SERVICE_NAME}.log"

# 创建 bin 目录
mkdir -p $BIN_DIR

# 构建可执行文件
echo "Building $SERVICE_NAME..."
go build -o $BIN_DIR/$SERVICE_NAME .

# 检查构建是否成功
if [ $? -ne 0 ]; then
    echo "Build failed. Exiting."
    exit 1
fi

# 启动服务后台运行，日志带时间戳
echo "Starting $SERVICE_NAME..."
nohup stdbuf -oL -eL $BIN_DIR/$SERVICE_NAME \
    > >(awk '{ print strftime("[%Y-%m-%d %H:%M:%S]"), $0; fflush(); }' >> $LOG_FILE) \
    2> >(awk '{ print strftime("[%Y-%m-%d %H:%M:%S] [ERR]"), $0; fflush(); }' >> $LOG_FILE) &

# 保存 PID
echo $! > $BIN_DIR/${SERVICE_NAME}.pid
echo "$SERVICE_NAME started with PID $(cat $BIN_DIR/${SERVICE_NAME}.pid), logs at $LOG_FILE"
