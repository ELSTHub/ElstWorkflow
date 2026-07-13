# Context Package

context 包提供了线程安全的工作流上下文实现，用于在工作流节点之间传递数据。

## 特性

- **线程安全**: 使用读写锁保证并发安全
- **类型安全**: 支持泛型的 GetT 方法
- **完整操作**: 支持 Put, Get, Delete, Clone, Keys 等操作
- **无业务依赖**: 通用键值存储，不包含业务特定字段

## 主要接口

### Context 接口

```go
type Context interface {
    Put(key string, value interface{})
    Get(key string) (interface{}, bool)
    GetTyped(key string, target interface{}) (interface{}, error)
    Delete(key string)
    Keys() []string
    Clone() Context
    Has(key string) bool
    Len() int
    Clear()
}
```

### 泛型辅助函数

```go
// GetT 泛型版本的Get方法，提供类型安全的值获取
func GetT[T any](ctx core.Context, key string) (T, error)

// MustGetT 泛型版本的Get方法，如果键不存在或类型不匹配会panic
func MustGetT[T any](ctx core.Context, key string) T
```

## 使用示例

### 基本使用

```go
ctx := context.New()

// 存储值
ctx.Put("name", "workflow-1")
ctx.Put("count", 42)

// 获取值
name, ok := ctx.Get("name")
if ok {
    fmt.Println(name) // "workflow-1"
}

// 泛型获取
count, err := context.GetT[int](ctx, "count")
if err == nil {
    fmt.Println(count) // 42
}
```

### 克隆上下文

```go
original := context.New()
original.Put("key", "value")

cloned := original.Clone()

// 修改原始不影响克隆
original.Put("new_key", "new_value")

// 修改克隆不影响原始
cloned.Put("cloned_key", "cloned_value")
```

### 并发使用

```go
ctx := context.New()
var wg sync.WaitGroup

// 并发写入
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()
        ctx.Put(fmt.Sprintf("key%d", n), n)
    }(i)
}

wg.Wait()
fmt.Println(ctx.Len()) // 100
```

## 设计原则

1. **线程安全**: 所有操作都使用读写锁保护
2. **接口隔离**: 实现 core.Context 接口，不暴露内部实现
3. **泛型支持**: 通过辅助函数提供类型安全的访问
4. **无状态**: 上下文不包含任何工作流状态信息

## 性能考虑

- 读操作使用读锁，支持并发读取
- 写操作使用写锁，保证数据一致性
- 克隆操作是深拷贝，修改克隆不影响原始
