// DAG 高级示例演示了复杂 DAG 工作流的并行执行
package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ELSTHub/elstworkflow/builder"
	"github.com/ELSTHub/elstworkflow/context"
	"github.com/ELSTHub/elstworkflow/core"
	"github.com/ELSTHub/elstworkflow/engine"
)

// 并行任务执行统计
var (
	mu      sync.Mutex
	running = make(map[string]bool)
	maxConc = 0
	curConc = 0
)

func trackExecution(name string) {
	mu.Lock()
	running[name] = true
	curConc++
	if curConc > maxConc {
		maxConc = curConc
	}
	mu.Unlock()

	fmt.Printf("  [%s] 开始执行 (当前并发: %d)\n", name, curConc)
}

func finishExecution(name string, duration time.Duration) {
	mu.Lock()
	delete(running, name)
	curConc--
	mu.Unlock()

	fmt.Printf("  [%s] 执行完成 (耗时: %v)\n", name, duration)
}

func main() {
	fmt.Println("=== DAG 高级示例：并行执行分析 ===")
	fmt.Println()

	// 构建复杂的 DAG 工作流
	//
	// 工作流结构:
	//
	//        ┌─── B ───┐
	//        │         │
	//   A ───┼─── C ───┼─── F ─── H
	//        │         │
	//        └─── D ───┘
	//              │
	//              E ─── G
	//
	// 说明:
	// - B, C, D 可以并行执行（都依赖 A）
	// - F 依赖 B, C, D
	// - E 依赖 D
	// - G 依赖 E
	// - H 依赖 F, G

	wf, err := builder.New("dag-advanced").
		WithVersion("1.0.0").
		WithDescription("复杂 DAG 工作流示例").
		// 节点 A: 数据准备
		Node("A", func(ctx core.Context) (interface{}, error) {
			trackExecution("A")
			time.Sleep(50 * time.Millisecond)
			finishExecution("A", 50*time.Millisecond)
			return "data-from-A", nil
		}).
		// 节点 B: 数据处理分支1
		Node("B", func(ctx core.Context) (interface{}, error) {
			trackExecution("B")
			time.Sleep(100 * time.Millisecond)
			finishExecution("B", 100*time.Millisecond)
			return "data-from-B", nil
		}).
		// 节点 C: 数据处理分支2
		Node("C", func(ctx core.Context) (interface{}, error) {
			trackExecution("C")
			time.Sleep(80 * time.Millisecond)
			finishExecution("C", 80*time.Millisecond)
			return "data-from-C", nil
		}).
		// 节点 D: 数据处理分支3
		Node("D", func(ctx core.Context) (interface{}, error) {
			trackExecution("D")
			time.Sleep(60 * time.Millisecond)
			finishExecution("D", 60*time.Millisecond)
			return "data-from-D", nil
		}).
		// 节点 E: D 的后续处理
		Node("E", func(ctx core.Context) (interface{}, error) {
			trackExecution("E")
			time.Sleep(70 * time.Millisecond)
			finishExecution("E", 70*time.Millisecond)
			return "data-from-E", nil
		}).
		// 节点 F: 合并 B, C, D 的结果
		Node("F", func(ctx core.Context) (interface{}, error) {
			trackExecution("F")
			b, _ := context.GetT[string](ctx, "B")
			c, _ := context.GetT[string](ctx, "C")
			d, _ := context.GetT[string](ctx, "D")
			fmt.Printf("    F 合并: %s, %s, %s\n", b, c, d)
			time.Sleep(90 * time.Millisecond)
			finishExecution("F", 90*time.Millisecond)
			return "data-from-F", nil
		}).
		// 节点 G: E 的后续处理
		Node("G", func(ctx core.Context) (interface{}, error) {
			trackExecution("G")
			time.Sleep(40 * time.Millisecond)
			finishExecution("G", 40*time.Millisecond)
			return "data-from-G", nil
		}).
		// 节点 H: 最终合并
		Node("H", func(ctx core.Context) (interface{}, error) {
			trackExecution("H")
			f, _ := context.GetT[string](ctx, "F")
			g, _ := context.GetT[string](ctx, "G")
			fmt.Printf("    H 最终合并: %s, %s\n", f, g)
			time.Sleep(30 * time.Millisecond)
			finishExecution("H", 30*time.Millisecond)
			return "final-result", nil
		}).
		// 定义依赖关系
		DependsOn("B", "A").
		DependsOn("C", "A").
		DependsOn("D", "A").
		DependsOn("E", "D").
		DependsOn("F", "B", "C", "D").
		DependsOn("G", "E").
		DependsOn("H", "F", "G").
		// 标记可并行执行的节点
		Parallel("B").
		Parallel("C").
		Parallel("D").
		Build()

	if err != nil {
		log.Fatalf("构建工作流失败: %v", err)
	}

	// 使用 DAG 调度器，支持并行执行
	config := &engine.Config{
		MaxParallel:   4, // 最大并行数
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
	fmt.Println("开始执行 DAG 工作流...")
	fmt.Println("════════════════════════════════════")

	start := time.Now()
	result, err := e.Run(ctx)
	if err != nil {
		log.Fatalf("运行工作流失败: %v", err)
	}
	totalDuration := time.Since(start)

	fmt.Println("════════════════════════════════════")

	// 输出结果
	fmt.Println("\n执行统计:")
	fmt.Printf("  总耗时: %v\n", totalDuration)
	fmt.Printf("  最大并发数: %d\n", maxConc)
	fmt.Printf("  执行节点数: %d\n", len(result.NodeResults))

	// 计算串行执行时间
	var serialDuration time.Duration
	for _, nodeResult := range result.NodeResults {
		serialDuration += nodeResult.Duration
	}

	fmt.Printf("  串行执行时间: %v\n", serialDuration)
	fmt.Printf("  加速比: %.2fx\n", float64(serialDuration)/float64(totalDuration))

	fmt.Println("\n节点执行顺序:")
	// 按时间排序输出节点执行情况
	type nodeInfo struct {
		name     string
		start    time.Time
		duration time.Duration
	}

	nodes := make([]nodeInfo, 0, len(result.NodeResults))
	for name, nr := range result.NodeResults {
		nodes = append(nodes, nodeInfo{
			name:     name,
			start:    nr.StartTime,
			duration: nr.Duration,
		})
	}

	// 简单排序
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[j].start.Before(nodes[i].start) {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}

	for _, n := range nodes {
		fmt.Printf("  %s: 开始=%v, 耗时=%v\n",
			n.name,
			n.start.Sub(start),
			n.duration,
		)
	}

	fmt.Println("\nDAG 拓扑结构:")
	fmt.Println("     ┌─── B ───┐")
	fmt.Println("     │         │")
	fmt.Println("A ───┼─── C ───┼─── F ─── H")
	fmt.Println("     │         │")
	fmt.Println("     └─── D ───┘")
	fmt.Println("           │")
	fmt.Println("           E ─── G")
}
