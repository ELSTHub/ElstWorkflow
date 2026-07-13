// Saga + DAG 组合示例
// 演示分布式事务与并行执行的结合使用
package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/elstworkflow/builder"
	"github.com/elstworkflow/context"
	"github.com/elstworkflow/core"
	"github.com/elstworkflow/engine"
	"github.com/elstworkflow/rollback"
)

func main() {
	fmt.Println("═════════════════════════════════════════════════════")
	fmt.Println("    Saga + DAG 组合示例：电商订单处理系统")
	fmt.Println("═════════════════════════════════════════════════════")
	fmt.Println()

	// 创建回滚管理器
	rollbackMgr := rollback.NewRollbackManager()

	// 构建工作流
	//
	// 结构说明:
	// - create-order, reduce-inventory, process-payment: Saga 事务节点（带回滚）
	// - send-email, send-sms, update-analytics: DAG 并行节点
	//
	//              ┌─── send-email ───┐
	//              │                  │
	// create-order → reduce-inventory → process-payment ─┼─── send-sms ─────┼─── update-analytics
	//              │                  │
	//              └──────────────────┘
	//
	// 如果任何 Saga 节点失败，会触发回滚
	// 如果 Saga 成功，通知节点可以并行执行

	wf, err := builder.New("ecommerce-order").
		WithVersion("1.0.0").
		WithDescription("电商订单处理系统").

		// ========== Saga 事务节点 ==========

		// 步骤1: 创建订单
		NodeWithRollback("create-order",
			func(ctx core.Context) (interface{}, error) {
				orderID := fmt.Sprintf("ORD-%d", time.Now().UnixNano()%10000)
				fmt.Printf("  ✓ [Saga] 创建订单: %s\n", orderID)
				ctx.Put("order-id", orderID)
				time.Sleep(50 * time.Millisecond)
				return orderID, nil
			},
			func(ctx core.Context) error {
				orderID, _ := context.GetT[string](ctx, "order-id")
				fmt.Printf("  ✗ [补偿] 取消订单: %s\n", orderID)
				return nil
			},
		).

		// 步骤2: 扣减库存
		NodeWithRollback("reduce-inventory",
			func(ctx core.Context) (interface{}, error) {
				orderID, _ := context.GetT[string](ctx, "order-id")
				fmt.Printf("  ✓ [Saga] 扣减库存 (订单: %s)\n", orderID)
				ctx.Put("inventory-reduced", true)
				time.Sleep(30 * time.Millisecond)
				return true, nil
			},
			func(ctx core.Context) error {
				orderID, _ := context.GetT[string](ctx, "order-id")
				fmt.Printf("  ✗ [补偿] 恢复库存 (订单: %s)\n", orderID)
				return nil
			},
		).

		// 步骤3: 处理支付
		NodeWithRollback("process-payment",
			func(ctx core.Context) (interface{}, error) {
				orderID, _ := context.GetT[string](ctx, "order-id")
				paymentID := fmt.Sprintf("PAY-%d", time.Now().UnixNano()%10000)
				fmt.Printf("  ✓ [Saga] 处理支付: %s (订单: %s)\n", paymentID, orderID)
				ctx.Put("payment-id", paymentID)
				time.Sleep(40 * time.Millisecond)

				// 模拟支付失败场景（取消注释以测试回滚）
				// return nil, errors.New("支付失败")

				return paymentID, nil
			},
			func(ctx core.Context) error {
				paymentID, _ := context.GetT[string](ctx, "payment-id")
				fmt.Printf("  ✗ [补偿] 退款: %s\n", paymentID)
				return nil
			},
		).

		// ========== DAG 并行节点 ==========

		// 发送邮件通知
		Node("send-email", func(ctx core.Context) (interface{}, error) {
			orderID, _ := context.GetT[string](ctx, "order-id")
			fmt.Printf("  → [并行] 发送邮件通知 (订单: %s)\n", orderID)
			time.Sleep(20 * time.Millisecond)
			return "email-sent", nil
		}).

		// 发送短信通知
		Node("send-sms", func(ctx core.Context) (interface{}, error) {
			orderID, _ := context.GetT[string](ctx, "order-id")
			fmt.Printf("  → [并行] 发送短信通知 (订单: %s)\n", orderID)
			time.Sleep(15 * time.Millisecond)
			return "sms-sent", nil
		}).

		// 更新分析数据
		Node("update-analytics", func(ctx core.Context) (interface{}, error) {
			orderID, _ := context.GetT[string](ctx, "order-id")
			fmt.Printf("  → [并行] 更新分析数据 (订单: %s)\n", orderID)
			time.Sleep(25 * time.Millisecond)
			return "analytics-updated", nil
		}).

		// 最终确认
		Node("confirm-order", func(ctx core.Context) (interface{}, error) {
			orderID, _ := context.GetT[string](ctx, "order-id")
			email, _ := context.GetT[string](ctx, "send-email")
			sms, _ := context.GetT[string](ctx, "send-sms")
			analytics, _ := context.GetT[string](ctx, "update-analytics")
			fmt.Printf("  ✓ [完成] 订单确认: %s (邮件:%s, 短信:%s, 分析:%s)\n",
				orderID, email, sms, analytics)
			return "confirmed", nil
		}).

		// ========== 依赖关系 ==========

		// Saga 事务依赖
		DependsOn("reduce-inventory", "create-order").
		DependsOn("process-payment", "reduce-inventory").

		// DAG 并行依赖
		DependsOn("send-email", "process-payment").
		DependsOn("send-sms", "process-payment").
		DependsOn("update-analytics", "process-payment").

		// 最终确认依赖所有通知完成
		DependsOn("confirm-order", "send-email", "send-sms", "update-analytics").

		// 标记可并行执行
		Parallel("send-email").
		Parallel("send-sms").
		Parallel("update-analytics").
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
	fmt.Println("开始执行订单处理工作流...")
	fmt.Println("─────────────────────────────────────────────────────")

	start := time.Now()
	result, err := e.Run(ctx)
	duration := time.Since(start)

	fmt.Println("─────────────────────────────────────────────────────")

	if err != nil {
		fmt.Printf("\n❌ 工作流执行失败: %v\n", err)
		fmt.Println("\n开始执行 Saga 补偿（回滚）...")
		fmt.Println("─────────────────────────────────────────────────────")

		// 执行回滚
		if rollbackErr := rollbackMgr.Rollback(ctx); rollbackErr != nil {
			fmt.Printf("回滚失败: %v\n", rollbackErr)
		} else {
			fmt.Println("─────────────────────────────────────────────────────")
			fmt.Println("✅ Saga 补偿完成")
		}
	}

	// 输出结果
	fmt.Println("\n═════════════════════════════════════════════════════")
	fmt.Println("执行结果:")
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Printf("  状态: %v\n", result.Status)
	fmt.Printf("  总耗时: %v\n", duration)
	fmt.Printf("  执行节点数: %d\n", len(result.NodeResults))

	fmt.Println("\n节点执行详情:")
	fmt.Println("─────────────────────────────────────────────────────")

	// 分类统计
	sagaNodes := []string{"create-order", "reduce-inventory", "process-payment"}
	dagNodes := []string{"send-email", "send-sms", "update-analytics", "confirm-order"}

	fmt.Println("  Saga 事务节点:")
	for _, name := range sagaNodes {
		if nr, ok := result.NodeResults[name]; ok {
			status := "✓"
			if nr.Status == core.NodeFailed {
				status = "✗"
			}
			fmt.Printf("    %s %s: %v\n", status, name, nr.Duration)
		}
	}

	fmt.Println("  DAG 并行节点:")
	for _, name := range dagNodes {
		if nr, ok := result.NodeResults[name]; ok {
			status := "✓"
			if nr.Status == core.NodeFailed {
				status = "✗"
			}
			fmt.Printf("    %s %s: %v\n", status, name, nr.Duration)
		}
	}

	fmt.Println("\n═════════════════════════════════════════════════════")
	fmt.Println("模式说明:")
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println("  • Saga 模式: 保证事务一致性，失败时自动回滚")
	fmt.Println("  • DAG 模式: 支持并行执行，提高处理效率")
	fmt.Println("  • 组合使用: 先 Saga 保证数据一致，后 DAG 并行处理")
	fmt.Println("═════════════════════════════════════════════════════")

	// 演示失败场景
	fmt.Println("\n\n=== 演示失败场景（支付失败触发回滚）===")
	fmt.Println()

	// 构建失败场景的工作流
	wfFail, _ := builder.New("ecommerce-order-fail").
		NodeWithRollback("create-order",
			func(ctx core.Context) (interface{}, error) {
				fmt.Println("  ✓ 创建订单: ORD-FAIL")
				ctx.Put("order-id", "ORD-FAIL")
				return "ORD-FAIL", nil
			},
			func(ctx core.Context) error {
				fmt.Println("  ✗ 回滚: 取消订单 ORD-FAIL")
				return nil
			},
		).
		NodeWithRollback("process-payment",
			func(ctx core.Context) (interface{}, error) {
				fmt.Println("  ✗ 支付失败!")
				return nil, errors.New("余额不足")
			},
			func(ctx core.Context) error {
				fmt.Println("  ✗ 回滚: 无需退款（支付未成功）")
				return nil
			},
		).
		DependsOn("process-payment", "create-order").
		Build()

	e2 := engine.New(nil)
	e2.Load(wfFail)

	fmt.Println("执行失败场景工作流...")
	fmt.Println("─────────────────────────────────────────────────────")

	result2, err2 := e2.Run(context.New())
	if err2 != nil {
		fmt.Printf("\n❌ 工作流失败: %v\n", err2)
		fmt.Println("\n执行 Saga 补偿...")
		rollbackMgr.Rollback(context.New())
	}

	fmt.Printf("\n最终状态: %v\n", result2.Status)
}
