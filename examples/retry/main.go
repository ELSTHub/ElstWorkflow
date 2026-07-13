// Retry 示例演示了重试策略的使用
package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ELSTHub/elstworkflow/builder"
	"github.com/ELSTHub/elstworkflow/context"
	"github.com/ELSTHub/elstworkflow/core"
	"github.com/ELSTHub/elstworkflow/engine"
)

func main() {
	attempts := 0

	// 构建带重试的工作流
	wf, err := builder.New("retry-workflow").
		WithVersion("1.0.0").
		WithDescription("一个带重试策略的工作流示例").
		Node("unstable-task", func(ctx core.Context) (interface{}, error) {
			attempts++
			fmt.Printf("尝试第 %d 次执行不稳定任务...\n", attempts)

			// 模拟前两次失败，第三次成功
			if attempts < 3 {
				return nil, errors.New("暂时性错误")
			}

			return "success", nil
		}).
		WithRetryPolicy("unstable-task", &core.RetryPolicy{
			Strategy:   core.RetryFixed,
			MaxRetries: 3,
			Interval:   100 * time.Millisecond,
		}).
		Node("finalize", func(ctx core.Context) (interface{}, error) {
			result, _ := context.GetT[string](ctx, "unstable-task")
			fmt.Printf("最终处理: %s\n", result)
			return "done", nil
		}).
		DependsOn("finalize", "unstable-task").
		Build()

	if err != nil {
		log.Fatalf("构建工作流失败: %v", err)
	}

	// 创建引擎
	e := engine.New(nil)

	// 加载工作流
	if err := e.Load(wf); err != nil {
		log.Fatalf("加载工作流失败: %v", err)
	}

	// 创建上下文
	ctx := context.New()

	// 运行工作流
	fmt.Println("开始执行带重试的工作流...")
	result, err := e.Run(ctx)
	if err != nil {
		log.Fatalf("运行工作流失败: %v", err)
	}

	// 输出结果
	fmt.Printf("\n工作流执行完成!\n")
	fmt.Printf("状态: %v\n", result.Status)
	fmt.Printf("总尝试次数: %d\n", attempts)
	fmt.Printf("耗时: %v\n", result.Duration)
}
