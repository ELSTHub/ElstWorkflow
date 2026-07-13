// Serial 示例演示了串行工作流的使用
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
	// 构建串行工作流
	wf, err := builder.New("serial-workflow").
		WithVersion("1.0.0").
		WithDescription("一个简单的串行工作流示例").
		Node("step1", func(ctx core.Context) (interface{}, error) {
			fmt.Println("执行步骤1: 初始化")
			return "step1-result", nil
		}).
		Node("step2", func(ctx core.Context) (interface{}, error) {
			// 获取step1的结果
			result, _ := context.GetT[string](ctx, "step1")
			fmt.Printf("执行步骤2: 处理 %s\n", result)
			return "step2-result", nil
		}).
		Node("step3", func(ctx core.Context) (interface{}, error) {
			// 获取step2的结果
			result, _ := context.GetT[string](ctx, "step2")
			fmt.Printf("执行步骤3: 完成 %s\n", result)
			return "step3-result", nil
		}).
		DependsOn("step2", "step1").
		DependsOn("step3", "step2").
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
	fmt.Println("开始执行串行工作流...")
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
