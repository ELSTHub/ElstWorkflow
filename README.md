# Go Workflow Framework

一个完全独立、可复用、生产可用的 Go Workflow Framework。

## 特性

- **声明式设计**: 使用 DAG 描述工作流，不依赖函数调用顺序
- **模块化架构**: 所有模块通过接口设计，可独立替换
- **零外部依赖**: 不依赖任何第三方库
- **线程安全**: 所有组件都支持并发访问
- **丰富的功能**:
  - DAG 调度
  - 并行执行
  - 重试策略（固定间隔、指数退避、自定义）
  - 超时控制
  - Saga 模式回滚
  - 检查点
  - 事件总线
  - 中间件支持
  - 节点注册表
  - JSON/YAML 编解码

## 快速开始

### 安装

```bash
go get github.com/elstworkflow
```

### 基本使用

```go
package main

import (
	"fmt"
	"log"

	"github.com/elstworkflow/builder"
	"github.com/elstworkflow/context"
	"github.com/elstworkflow/core"
	"github.com/elstworkflow/engine"
)

func main() {
	// 构建工作流
	wf, err := builder.New("my-workflow").
		Node("step1", func(ctx core.Context) (interface{}, error) {
			return "result1", nil
		}).
		Node("step2", func(ctx core.Context) (interface{}, error) {
			return "result2", nil
		}).
		DependsOn("step2", "step1").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// 创建并运行引擎
	e := engine.New(nil)
	e.Load(wf)

	result, err := e.Run(context.New())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("状态: %v\n", result.Status)
}
```

## 包结构

| 包 | 描述 |
|---|---|
| `core` | 核心类型和接口 |
| `context` | 线程安全的工作流上下文 |
| `node` | 节点接口和默认实现 |
| `graph` | DAG 图结构 |
| `builder` | 工作流构建器 |
| `scheduler` | 调度器（串行、DAG、优先级） |
| `executor` | 节点执行器 |
| `engine` | 工作流引擎 |
| `retry` | 重试策略 |
| `timeout` | 超时策略 |
| `rollback` | Saga 模式回滚 |
| `event` | 事件总线 |
| `middleware` | 中间件 |
| `registry` | 节点注册表 |
| `codec` | JSON/YAML 编解码 |
| `persistence` | 持久化存储 |

## 示例

- [serial](examples/serial/): 串行工作流
- [parallel](examples/parallel/): 并行工作流
- [dag](examples/dag/): 复杂 DAG 工作流
- [retry](examples/retry/): 重试策略
- [rollback](examples/rollback/): Saga 模式回滚

## 设计原则

1. **接口优先**: 所有模块通过接口定义，实现可替换
2. **无业务依赖**: 框架本身不包含任何业务逻辑
3. **声明式**: 使用 DAG 描述工作流结构
4. **可扩展**: 支持自定义节点、中间件、存储等

## 许可证

MIT License
