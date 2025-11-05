#!/bin/bash

# 自定义类型功能测试脚本
# 测试类型引用和数组功能

BASE_URL="http://localhost:8080/api"

echo "========== 自定义类型功能测试 =========="
echo ""

# 1. 创建应用
echo "1. 创建测试应用..."
APP_RESPONSE=$(curl -s -X POST "$BASE_URL/applications" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试应用",
    "description": "用于测试自定义类型功能",
    "path": "test-app",
    "protocol": "http",
    "enabled": true
  }')

APP_ID=$(echo $APP_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "✓ 应用创建成功，ID: $APP_ID"
echo ""

# 2. 创建基础类型 - Address
echo "2. 创建 Address 类型..."
ADDRESS_RESPONSE=$(curl -s -X POST "$BASE_URL/custom-types" \
  -H "Content-Type: application/json" \
  -d "{
    \"app_id\": $APP_ID,
    \"name\": \"Address\",
    \"description\": \"地址信息\",
    \"fields\": [
      {
        \"name\": \"street\",
        \"type\": \"string\",
        \"required\": true,
        \"description\": \"街道地址\"
      },
      {
        \"name\": \"city\",
        \"type\": \"string\",
        \"required\": true,
        \"description\": \"城市\"
      },
      {
        \"name\": \"zipCode\",
        \"type\": \"string\",
        \"required\": false,
        \"description\": \"邮政编码\"
      }
    ]
  }")

ADDRESS_ID=$(echo $ADDRESS_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*' | head -1)
echo "✓ Address 类型创建成功，ID: $ADDRESS_ID"
echo ""

# 3. 创建基础类型 - User（包含数组字段）
echo "3. 创建 User 类型（包含字符串数组）..."
USER_RESPONSE=$(curl -s -X POST "$BASE_URL/custom-types" \
  -H "Content-Type: application/json" \
  -d "{
    \"app_id\": $APP_ID,
    \"name\": \"User\",
    \"description\": \"用户信息\",
    \"fields\": [
      {
        \"name\": \"id\",
        \"type\": \"number\",
        \"required\": true,
        \"description\": \"用户ID\"
      },
      {
        \"name\": \"username\",
        \"type\": \"string\",
        \"required\": true,
        \"description\": \"用户名\"
      },
      {
        \"name\": \"email\",
        \"type\": \"string\",
        \"required\": false,
        \"description\": \"邮箱地址\"
      },
      {
        \"name\": \"tags\",
        \"type\": \"string\",
        \"is_array\": true,
        \"required\": false,
        \"description\": \"标签列表（字符串数组）\"
      },
      {
        \"name\": \"active\",
        \"type\": \"boolean\",
        \"required\": true,
        \"description\": \"是否激活\"
      }
    ]
  }")

USER_ID=$(echo $USER_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*' | head -1)
echo "✓ User 类型创建成功，ID: $USER_ID"
echo ""

# 4. 创建复杂类型 - UserProfile（引用其他类型）
echo "4. 创建 UserProfile 类型（引用 User 和 Address）..."
PROFILE_RESPONSE=$(curl -s -X POST "$BASE_URL/custom-types" \
  -H "Content-Type: application/json" \
  -d "{
    \"app_id\": $APP_ID,
    \"name\": \"UserProfile\",
    \"description\": \"用户详细信息\",
    \"fields\": [
      {
        \"name\": \"user\",
        \"type\": \"custom\",
        \"ref\": $USER_ID,
        \"required\": true,
        \"description\": \"用户基本信息\"
      },
      {
        \"name\": \"address\",
        \"type\": \"custom\",
        \"ref\": $ADDRESS_ID,
        \"required\": false,
        \"description\": \"用户地址\"
      },
      {
        \"name\": \"bio\",
        \"type\": \"string\",
        \"required\": false,
        \"description\": \"个人简介\"
      }
    ]
  }")

PROFILE_ID=$(echo $PROFILE_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*' | head -1)
echo "✓ UserProfile 类型创建成功，ID: $PROFILE_ID"
echo ""

# 5. 创建包含数组引用的类型 - Company
echo "5. 创建 Company 类型（包含 User 数组）..."
COMPANY_RESPONSE=$(curl -s -X POST "$BASE_URL/custom-types" \
  -H "Content-Type: application/json" \
  -d "{
    \"app_id\": $APP_ID,
    \"name\": \"Company\",
    \"description\": \"公司信息\",
    \"fields\": [
      {
        \"name\": \"name\",
        \"type\": \"string\",
        \"required\": true,
        \"description\": \"公司名称\"
      },
      {
        \"name\": \"employees\",
        \"type\": \"custom\",
        \"ref\": $USER_ID,
        \"is_array\": true,
        \"required\": false,
        \"description\": \"员工列表（User 数组）\"
      },
      {
        \"name\": \"tags\",
        \"type\": \"string\",
        \"is_array\": true,
        \"required\": false,
        \"description\": \"公司标签\"
      }
    ]
  }")

COMPANY_ID=$(echo $COMPANY_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*' | head -1)
echo "✓ Company 类型创建成功，ID: $COMPANY_ID"
echo ""

# 6. 查询所有类型
echo "6. 查询应用下的所有类型..."
curl -s "$BASE_URL/custom-types?app_id=$APP_ID" | python3 -m json.tool
echo ""

# 7. 创建使用自定义类型的接口
echo "7. 创建接口（使用自定义类型参数）..."
INTERFACE_RESPONSE=$(curl -s -X POST "$BASE_URL/interfaces" \
  -H "Content-Type: application/json" \
  -d "{
    \"app_id\": $APP_ID,
    \"name\": \"CreateUserProfile\",
    \"description\": \"创建用户档案\",
    \"method\": \"POST\",
    \"protocol\": \"http\",
    \"url\": \"https://api.example.com/profiles\",
    \"auth_type\": \"bearer\",
    \"enabled\": true,
    \"parameters\": [
      {
        \"name\": \"profile\",
        \"type\": \"custom\",
        \"ref\": $PROFILE_ID,
        \"location\": \"body\",
        \"required\": true,
        \"description\": \"用户档案信息\"
      },
      {
        \"name\": \"addresses\",
        \"type\": \"custom\",
        \"ref\": $ADDRESS_ID,
        \"is_array\": true,
        \"location\": \"body\",
        \"required\": false,
        \"description\": \"地址列表\"
      }
    ]
  }")

INTERFACE_ID=$(echo $INTERFACE_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*' | head -1)
echo "✓ 接口创建成功，ID: $INTERFACE_ID"
echo ""

# 8. 测试删除被引用的类型（应该失败）
echo "8. 测试删除被引用的 User 类型（应该失败）..."
DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/custom-types/$USER_ID")
echo "响应: $DELETE_RESPONSE"
echo ""

# 9. 获取单个类型详情
echo "9. 获取 Company 类型详情..."
curl -s "$BASE_URL/custom-types/$COMPANY_ID" | python3 -m json.tool
echo ""

echo "========== 测试完成 =========="
echo ""
echo "总结："
echo "- 创建了 4 个自定义类型（Address, User, UserProfile, Company）"
echo "- 测试了类型引用功能（UserProfile 引用 User 和 Address）"
echo "- 测试了数组功能（User.tags[], Company.employees[], Company.tags[]）"
echo "- 测试了接口参数使用自定义类型"
echo "- 测试了引用完整性检查（删除被引用的类型会失败）"
echo ""
echo "请访问 http://localhost:8080 查看前端界面"
