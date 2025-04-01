#!/bin/bash

bash ./stop.sh

if [ $? -eq 0 ]; then
    bash ./start.sh
else
    echo "重启应用失败。"
    exit 1
fi
