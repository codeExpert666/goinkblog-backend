#!/bin/bash

# 定义应用名称
APP_NAME=goinkblog

# 获取脚本所在目录
SCRIPT_DIR=$(dirname "$0")

# 导航到项目根目录（脚本目录的上一级）
cd "$SCRIPT_DIR/.."

# 停止应用
echo "正在停止应用..."
./$APP_NAME stop

# 检查是否停止成功
if [ $? -ne 0 ]; then
    echo "停止应用失败。"
    exit 1
fi

echo "停止应用成功。"
