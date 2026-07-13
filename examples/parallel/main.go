// Parallel 示例演示了并行工作流的使用
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ELSTHub/elstworkflow/builder"
	"github.com/ELSTHub/elstworkflow/context"
	"github.com/ELSTHub/elstworkflow/core"
	"github.com/ELSTHub/elstworkflow/engine"
)

func main() {
	// 构建并行工作流
	wf, err := builder.New("parallel-workflow").
		WithVersion("1.0.0").
		WithDescription("一个并行工作流示例").
		Node("prepare", func(ctx core.Context) (interface{}, error) {
			fmt.Println("准备阶段: 初始化数据")
			time.Sleep(100 * time.Millisecond)
			return "prepared", nil
		}).
		Node("task1", func(ctx core.Context) (interface{}, error) {
			fmt.Println("任务1: 处理数据A")
			time.Sleep(200 * time.Millisecond)
			return "task1-result", nil
		}).
		Node("task2", func(ctx core.Context) (interface{}, error) {
			fmt.Println("任务2: 处理数据B")
			time.Sleep(150 * time.Millisecond)
			return "task2-result", nil
		}).
		Node("task3", func(ctx core.Context) (interface{}, error) {
			fmt.Println("任务3: 处理数据C")
			time.Sleep(100 * time.Millisecond)
			return "task3-result", nil
		}).
		Node("merge", func(ctx core.Context) (interface{}, error) {
			r1, _ := context.GetT[string](ctx, "task1")
			r2, _ := context.GetT[string](ctx, "task2")
			r3, _ := context.GetT[string](ctx, "task3")
			fmt.Printf("合并结果: %s, %s, %s\n", r1, r2, r3)
			return "merged", nil
		}).
		DependsOn("task1", "prepare").
		DependsOn("task2", "prepare").
		DependsOn("task3", "prepare").
		DependsOn("merge", "task1", "task2", "task3").
		Parallel("task1").
		Parallel("task2").
		Parallel("task3").
		Build()

	if err != nil {
		log.Fatalf("构建工作流失败: %v", err)
	}

	// 使用 DAG 调度器
	config := &engine.Config{
		MaxParallel:   3,
		SchedulerType: engine.DAGScheduler,
	}

	// 创建引擎
	e := engine.New(config)

	// 加载工作流
	if err := e.Load(wf); err != nil {
		log.Fatalf("加载工作流失败: %v", err)
	}

	// 创建上下文
	ctx := context.New()

	// 运行工作流
	fmt.Println("开始执行并行工作流...")
	start := time.Now()
	result, err := e.Run(ctx)
	if err != nil {
		log.Fatalf("运行工作流失败: %v", err)
	}

	// 输出结果
	fmt.Printf("\n工作流执行完成!\n")
	fmt.Printf("状态: %v\n", result.Status)
	fmt.Printf("总耗时: %v\n", time.Since(start))
	fmt.Printf("执行的节点数: %d\n", len(result.NodeResults))

	// 输出每个节点的结果
	for name, nodeResult := range result.NodeResults {
		fmt.Printf("节点 %s: 状态=%v, 耗时=%v\n", name, nodeResult.Status, nodeResult.Duration)
	}
}
