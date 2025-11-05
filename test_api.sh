#!/bin/bash

# API 测试脚本
# 用于验证重构后的 API 功能

BASE_URL="http://localhost:8080/api"

echo "========================================="
echo "MCP Adapter API 测试脚本"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试计数
TOTAL=0
PASSED=0
FAILED=0

# 测试函数
test_api() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    
    TOTAL=$((TOTAL + 1))
    echo -e "${YELLOW}测试 $TOTAL: $name${NC}"
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✓ 通过 (HTTP $http_code)${NC}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ 失败 (HTTP $http_code)${NC}"
        echo "$body"
        FAILED=$((FAILED + 1))
    fi
    echo ""
}

# 等待服务启动
echo "检查服务是否运行..."
if ! curl -s "$BASE_URL/applications" > /dev/null 2>&1; then
    echo -e "${RED}错误: 服务未运行，请先启动服务: go run main.go${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 服务正在运行${NC}"
echo ""

# 1. 创建应用
echo "========================================="
echo "1. 应用管理测试"
echo "========================================="
test_api "创建应用" "POST" "/applications" '{
  "name": "Test Application",
  "description": "测试应用",
  "path": "test-app",
  "protocol": "sse",
  "enabled": true
}'

test_api "获取应用列表" "GET" "/applications"

test_api "获取单个应用" "GET" "/applications/1"

# 2. 自定义类型测试
echo "========================================="
echo "2. 自定义类型测试"
echo "========================================="

test_api "创建 User 类型" "POST" "/custom-types" '{
  "app_id": 1,
  "name": "User",
  "description": "用户信息",
  "fields": [
    {
      "name": "id",
      "type": "number",
      "required": true,
      "description": "用户ID"
    },
    {
      "name": "username",
      "type": "string",
      "required": true,
      "description": "用户名"
    },
    {
      "name": "email",
      "type": "string",
      "required": false,
      "description": "邮箱"
    }
  ]
}'

test_api "创建 Address 类型" "POST" "/custom-types" '{
  "app_id": 1,
  "name": "Address",
  "description": "地址信息",
  "fields": [
    {
      "name": "city",
      "type": "string",
      "required": true,
      "description": "城市"
    },
    {
      "name": "street",
      "type": "string",
      "required": false,
      "description": "街道"
    }
  ]
}'

test_api "创建嵌套类型 UserProfile" "POST" "/custom-types" '{
  "app_id": 1,
  "name": "UserProfile",
  "description": "用户详细信息",
  "fields": [
    {
      "name": "user",
      "type": "custom",
      "ref": 1,
      "required": true,
      "description": "用户基本信息"
    },
    {
      "name": "address",
      "type": "custom",
      "ref": 2,
      "required": false,
      "description": "地址"
    },
    {
      "name": "tags",
      "type": "string",
      "is_array": true,
      "required": false,
      "description": "标签"
    }
  ]
}'

test_api "获取自定义类型列表" "GET" "/custom-types?app_id=1"

test_api "获取单个自定义类型" "GET" "/custom-types/1"

# 3. 接口测试
echo "========================================="
echo "3. 接口管理测试"
echo "========================================="

test_api "创建 GET 接口" "POST" "/interfaces" '{
  "app_id": 1,
  "name": "GetUser",
  "description": "获取用户信息",
  "protocol": "http",
  "url": "https://api.example.com/users",
  "method": "GET",
  "auth_type": "none",
  "enabled": true,
  "parameters": [
    {
      "name": "id",
      "type": "string",
      "location": "query",
      "required": true,
      "description": "用户ID"
    },
    {
      "name": "include_details",
      "type": "boolean",
      "location": "query",
      "required": false,
      "description": "是否包含详细信息",
      "default_value": "false"
    }
  ]
}'

test_api "创建 POST 接口（使用自定义类型）" "POST" "/interfaces" '{
  "app_id": 1,
  "name": "CreateUser",
  "description": "创建用户",
  "protocol": "http",
  "url": "https://api.example.com/users",
  "method": "POST",
  "auth_type": "none",
  "enabled": true,
  "parameters": [
    {
      "name": "user",
      "type": "custom",
      "ref": 1,
      "location": "body",
      "required": true,
      "description": "用户信息"
    },
    {
      "name": "Authorization",
      "type": "string",
      "location": "header",
      "required": true,
      "description": "认证令牌"
    }
  ]
}'

test_api "获取接口列表" "GET" "/interfaces?app_id=1"

test_api "获取单个接口" "GET" "/interfaces/1"

test_api "更新接口" "PUT" "/interfaces/1" '{
  "name": "GetUserById",
  "description": "根据ID获取用户（已更新）"
}'

# 4. 更新和删除测试
echo "========================================="
echo "4. 更新和删除测试"
echo "========================================="

test_api "更新自定义类型" "PUT" "/custom-types/1" '{
  "name": "UserInfo",
  "description": "用户基本信息（已更新）",
  "fields": [
    {
      "name": "id",
      "type": "number",
      "required": true
    },
    {
      "name": "username",
      "type": "string",
      "required": true
    },
    {
      "name": "nickname",
      "type": "string",
      "required": false,
      "description": "昵称（新增字段）"
    }
  ]
}'

test_api "更新应用" "PUT" "/applications/1" '{
  "description": "测试应用（已更新）"
}'

# 5. 错误处理测试
echo "========================================="
echo "5. 错误处理测试"
echo "========================================="

test_api "创建重复名称的类型（应失败）" "POST" "/custom-types" '{
  "app_id": 1,
  "name": "UserInfo",
  "fields": []
}'

test_api "引用不存在的类型（应失败）" "POST" "/custom-types" '{
  "app_id": 1,
  "name": "InvalidType",
  "fields": [
    {
      "name": "invalid",
      "type": "custom",
      "ref": 9999,
      "required": true
    }
  ]
}'

test_api "删除被引用的类型（应失败）" "DELETE" "/custom-types/1"

# 6. 清理测试（可选）
echo "========================================="
echo "6. 清理测试数据（可选）"
echo "========================================="
read -p "是否删除测试数据？(y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    test_api "删除接口 1" "DELETE" "/interfaces/1"
    test_api "删除接口 2" "DELETE" "/interfaces/2"
    test_api "删除类型 3" "DELETE" "/custom-types/3"
    test_api "删除类型 2" "DELETE" "/custom-types/2"
    test_api "删除类型 1" "DELETE" "/custom-types/1"
    test_api "删除应用" "DELETE" "/applications/1"
fi

# 测试总结
echo "========================================="
echo "测试总结"
echo "========================================="
echo -e "总计: $TOTAL"
echo -e "${GREEN}通过: $PASSED${NC}"
echo -e "${RED}失败: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ 所有测试通过！${NC}"
    exit 0
else
    echo -e "${RED}✗ 部分测试失败${NC}"
    exit 1
fi
