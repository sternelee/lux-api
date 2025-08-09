#!/bin/bash

# Lux API 测试脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查是否提供了API URL
if [ -z "$1" ]; then
    echo -e "${RED}❌ 请提供API URL作为参数${NC}"
    echo "用法: $0 <API_URL>"
    echo "示例: $0 https://xxxxxxxxxx.execute-api.region.amazonaws.com"
    exit 1
fi

API_URL="$1"

echo -e "${YELLOW}🧪 开始测试 Lux API...${NC}"
echo "API URL: $API_URL"
echo ""

# 测试GET请求
echo -e "${YELLOW}📡 测试 GET /extract...${NC}"
GET_RESPONSE=$(curl -s "$API_URL/extract?url=https://www.youtube.com/watch?v=dQw4w9WgXcQ")

if echo "$GET_RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✅ GET 请求成功${NC}"
    echo "响应: $GET_RESPONSE" | head -c 200
    echo "..."
else
    echo -e "${RED}❌ GET 请求失败${NC}"
    echo "响应: $GET_RESPONSE"
fi

echo ""

# 测试POST请求
echo -e "${YELLOW}📡 测试 POST /extract...${NC}"
POST_RESPONSE=$(curl -s -X POST "$API_URL/extract" \
    -H "Content-Type: application/json" \
    -d '{"url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ"}')

if echo "$POST_RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✅ POST 请求成功${NC}"
    echo "响应: $POST_RESPONSE" | head -c 200
    echo "..."
else
    echo -e "${RED}❌ POST 请求失败${NC}"
    echo "响应: $POST_RESPONSE"
fi

echo ""

# 测试错误情况
echo -e "${YELLOW}📡 测试错误情况（缺少URL参数）...${NC}"
ERROR_RESPONSE=$(curl -s "$API_URL/extract")

if echo "$ERROR_RESPONSE" | grep -q '"success":false'; then
    echo -e "${GREEN}✅ 错误处理正常${NC}"
    echo "响应: $ERROR_RESPONSE"
else
    echo -e "${RED}❌ 错误处理异常${NC}"
    echo "响应: $ERROR_RESPONSE"
fi

echo ""
echo -e "${GREEN}🎉 测试完成！${NC}"
