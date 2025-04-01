#!/bin/bash

# 定义应用名称 - 使用项目实际名称
APP_NAME=goinkblog

# 获取脚本所在目录
SCRIPT_DIR=$(dirname "$0")

# 导航到项目根目录（脚本目录的上一级）
cd "$SCRIPT_DIR/.."

# 构建应用
echo "正在构建应用..."
go build -o $APP_NAME

# 启动应用
echo "正在启动应用..."
./$APP_NAME start -d configs -c dev -s static -daemon

# 检查是否启动成功
if [ $? -ne 0 ]; then
    echo "启动应用失败。"
    exit 1
fi

echo "应用启动成功。"
