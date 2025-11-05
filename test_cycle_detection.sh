#!/bin/bash

# 测试循环引用检测功能
BASE_URL="http://localhost:8080"

echo -e "\033[36m=== 测试自定义类型循环引用检测 ===\033[0m"

# 1. 创建应用
echo -e "\n\033[33m1. 创建测试应用...\033[0m"
APP_RESPONSE=$(curl -s -X POST "$BASE_URL/api/applications" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "循环引用测试应用",
    "description": "测试DFS判环功能",
    "path": "cycle-test",
    "protocol": "sse",
    "enabled": true
  }')

APP_ID=$(echo $APP_RESPONSE | jq -r '.application.id')
echo -e "\033[32m应用创建成功, ID: $APP_ID\033[0m"

# 2. 创建类型A
echo -e "\n\033[33m2. 创建类型A...\033[0m"
TYPE_A_RESPONSE=$(curl -s -X POST "$BASE_URL/api/custom-types" \
  -H "Content-Type: application/json" \
  -d "{
    \"app_id\": $APP_ID,
    \"name\": \"TypeA\",
    \"description\": \"类型A\",
    \"fields\": [
      {
        \"name\": \"name\",
        \"type\": \"string\",
        \"required\": true,
        \"description\": \"名称\"
      }
    ]
  }")

TYPE_A_ID=$(echo $TYPE_A_RESPONSE | jq -r '.custom_type.id')
echo -e "\033[32m类型A创建成功, ID: $TYPE_A_ID\033[0m"

# 3. 创建类型B
echo -e "\n\033[33m3. 创建类型B...\033[0m"
TYPE_B_RESPONSE=$(curl -s -X POST "$BASE_URL/api/custom-types" \
  -H "Content-Type: application/json" \
  -d "{
    \"app_id\": $APP_ID,
    \"name\": \"TypeB\",
    \"description\": \"类型B\",
    \"fields\": [
      {
        \"name\": \"refA\",
        \"type\": \"custom\",
        \"ref\": $TYPE_A_ID,
        \"required\": false,
        \"description\": \"引用类型A\"
      }
    ]
  }")

TYPE_B_ID=$(echo $TYPE_B_RESPONSE | jq -r '.custom_type.id')
echo -e "\033[32m类型B创建成功, ID: $TYPE_B_ID\033[0m"

# 4. 尝试更新类型A,让它引用类型B (形成环: A -> B -> A)
echo -e "\n\033[33m4. 尝试更新类型A引用类型B (应该失败)...\033[0m"
UPDATE_RESPONSE=$(curl -s -X PUT "$BASE_URL/api/custom-types/$TYPE_A_ID" \
  -H "Content-Type: application/json" \
  -d "{
    \"fields\": [
      {
        \"name\": \"name\",
        \"type\": \"string\",
        \"required\": true,
        \"description\": \"名称\"
      },
      {
        \"name\": \"refB\",
        \"type\": \"custom\",
        \"ref\": $TYPE_B_ID,
        \"required\": false,
        \"description\": \"引用类型B\"
      }
    ]
  }")

if echo $UPDATE_RESPONSE | grep -q "circular reference"; then
  echo -e "\033[32m成功: 检测到循环引用!\033[0m"
  echo -e "\033[90m错误信息: $(echo $UPDATE_RESPONSE | jq -r '.error // .message')\033[0m"
else
  echo -e "\033[31m错误: 应该检测到循环引用但没有!\033[0m"
fi

# 5. 创建类型C (不形成环)
echo -e "\n\033[33m5. 创建类型C引用类型A (不形成环,应该成功)...\033[0m"
TYPE_C_RESPONSE=$(curl -s -X POST "$BASE_URL/api/custom-types" \
  -H "Content-Type: application/json" \
  -d "{
    \"app_id\": $APP_ID,
    \"name\": \"TypeC\",
    \"description\": \"类型C\",
    \"fields\": [
      {
        \"name\": \"refA\",
        \"type\": \"custom\",
        \"ref\": $TYPE_A_ID,
        \"required\": false,
        \"description\": \"引用类型A\"
      }
    ]
  }")

if echo $TYPE_C_RESPONSE | jq -e '.custom_type.id' > /dev/null 2>&1; then
  TYPE_C_ID=$(echo $TYPE_C_RESPONSE | jq -r '.custom_type.id')
  echo -e "\033[32m成功: 类型C创建成功, ID: $TYPE_C_ID\033[0m"
else
  echo -e "\033[31m错误: 类型C创建失败\033[0m"
fi

echo -e "\n\033[36m=== 测试接口参数循环引用检测 ===\033[0m"

# 6. 创建接口使用类型B
echo -e "\n\033[33m6. 创建接口使用类型B (应该成功)...\033[0m"
INTERFACE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/interfaces" \
  -H "Content-Type: application/json" \
  -d "{
    \"app_id\": $APP_ID,
    \"name\": \"TestInterface\",
    \"description\": \"测试接口\",
    \"protocol\": \"http\",
    \"url\": \"http://example.com/api/test\",
    \"method\": \"POST\",
    \"auth_type\": \"none\",
    \"enabled\": true,
    \"parameters\": [
      {
        \"name\": \"data\",
        \"type\": \"custom\",
        \"ref\": $TYPE_B_ID,
        \"location\": \"body\",
        \"required\": true,
        \"description\": \"使用类型B\"
      }
    ]
  }")

if echo $INTERFACE_RESPONSE | jq -e '.interface.id' > /dev/null 2>&1; then
  echo -e "\033[32m成功: 接口创建成功\033[0m"
else
  echo -e "\033[31m错误: 接口创建失败\033[0m"
fi

echo -e "\n\033[36m=== 测试完成 ===\033[0m"
echo -e "\033[33m清理测试数据...\033[0m"
curl -s -X DELETE "$BASE_URL/api/applications/$APP_ID" > /dev/null
echo -e "\033[32m测试应用已删除\033[0m"
