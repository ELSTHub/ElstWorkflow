// DAG 示例演示了复杂DAG工作流的使用
package main

import (
	"fmt"
	"log"

	"github.com/ELSTHub/elstworkflow/builder"
	"github.com/ELSTHub/elstworkflow/context"
	"github.com/ELSTHub/elstworkflow/core"
	"github.com/ELSTHub/elstworkflow/engine"
)

func main() {
	// 构建复杂的DAG工作流
	//
	// 工作流结构:
	//     A
	//    / \
	//   B   C
	//   |   |
	//   D   E
	//    \ /
	//     F
	//
	wf, err := builder.New("dag-workflow").
		WithVersion("1.0.0").
		WithDescription("一个复杂的DAG工作流示例").
		Node("A", func(ctx core.Context) (interface{}, error) {
			fmt.Println("节点A: 开始处理")
			return "A-done", nil
		}).
		Node("B", func(ctx core.Context) (interface{}, error) {
			a, _ := context.GetT[string](ctx, "A")
			fmt.Printf("节点B: 处理 %s 的结果\n", a)
			return "B-done", nil
		}).
		Node("C", func(ctx core.Context) (interface{}, error) {
			a, _ := context.GetT[string](ctx, "A")
			fmt.Printf("节点C: 处理 %s 的结果\n", a)
			return "C-done", nil
		}).
		Node("D", func(ctx core.Context) (interface{}, error) {
			b, _ := context.GetT[string](ctx, "B")
			fmt.Printf("节点D: 处理 %s 的结果\n", b)
			return "D-done", nil
		}).
		Node("E", func(ctx core.Context) (interface{}, error) {
			c, _ := context.GetT[string](ctx, "C")
			fmt.Printf("节点E: 处理 %s 的结果\n", c)
			return "E-done", nil
		}).
		Node("F", func(ctx core.Context) (interface{}, error) {
			d, _ := context.GetT[string](ctx, "D")
			e, _ := context.GetT[string](ctx, "E")
			fmt.Printf("节点F: 合并 %s 和 %s\n", d, e)
			return "F-done", nil
		}).
		DependsOn("B", "A").
		DependsOn("C", "A").
		DependsOn("D", "B").
		DependsOn("E", "C").
		DependsOn("F", "D", "E").
		Build()

	if err != nil {
		log.Fatalf("构建工作流失败: %v", err)
	}

	// 使用 DAG 调度器
	config := &engine.Config{
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
	fmt.Println("开始执行DAG工作流...")
	result, err := e.Run(ctx)
	if err != nil {
		log.Fatalf("运行工作流失败: %v", err)
	}

	// 输出结果
	fmt.Printf("\n工作流执行完成!\n")
	fmt.Printf("状态: %v\n", result.Status)
	fmt.Printf("耗时: %v\n", result.Duration)
	fmt.Printf("执行的节点数: %d\n", len(result.NodeResults))

	// 输出每个节点的结果
	for name, nodeResult := range result.NodeResults {
		fmt.Printf("节点 %s: 状态=%v, 输出=%v\n", name, nodeResult.Status, nodeResult.Output)
	}
}
