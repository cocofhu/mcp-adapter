# 循环引用检测 - 快速参考

## 什么是循环引用?

当类型 A 引用类型 B,而类型 B 又引用类型 A 时,就形成了循环引用。

```
❌ 错误示例:
TypeA { field: TypeB }
TypeB { field: TypeA }

❌ 自引用:
TypeA { field: TypeA }

✅ 正确示例:
TypeA { name: string }
TypeB { refA: TypeA }
TypeC { refB: TypeB }
```

## 何时会检测?

- ✅ 创建自定义类型时
- ✅ 更新自定义类型时
- ✅ 创建接口时
- ✅ 更新接口时

## 错误信息

如果检测到循环引用,会看到以下错误:

```
circular reference detected in custom type fields
```

或

```
circular reference detected in interface parameters
```

## 如何避免?

1. **规划类型层次**: 在设计类型时,确保引用关系是单向的
2. **使用基础类型**: 优先使用 `string`、`number`、`boolean` 等基础类型
3. **检查引用链**: 创建新类型前,检查是否会形成环

## 示例场景

### ✅ 允许: 树形结构
```
User { name: string }
Post { author: User }
Comment { post: Post }
```

### ❌ 禁止: 循环依赖
```
User { posts: Post[] }
Post { author: User, relatedPosts: Post[] }  // 如果 relatedPosts 引用回 User 会形成环
```

### ✅ 解决方案: 使用 ID 引用
```
User { name: string }
Post { 
  authorId: number,  // 使用 ID 而不是对象引用
  title: string 
}
```

## 测试

运行测试脚本验证功能:

```powershell
# Windows
.\test_cycle_detection.ps1

# Linux/Mac
./test_cycle_detection.sh
```

## 更多信息

详细文档请参考:
- [CYCLE_DETECTION.md](./CYCLE_DETECTION.md) - 完整功能说明
- [IMPLEMENTATION_SUMMARY.md](./IMPLEMENTATION_SUMMARY.md) - 实现细节
