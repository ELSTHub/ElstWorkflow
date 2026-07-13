// Saga 示例演示了完整的 Saga 模式分布式事务
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/ELSTHub/elstworkflow/builder"
	"github.com/ELSTHub/elstworkflow/context"
	"github.com/ELSTHub/elstworkflow/core"
	"github.com/ELSTHub/elstworkflow/engine"
	"github.com/ELSTHub/elstworkflow/rollback"
)

func main() {
	fmt.Println("=== Saga 模式分布式事务示例 ===")
	fmt.Println()

	// 创建回滚管理器
	rollbackMgr := rollback.NewRollbackManager()

	// 构建 Saga 工作流
	// 场景: 电商订单处理
	// 1. 创建订单 -> 2. 扣减库存 -> 3. 处理支付 -> 4. 发货
	// 如果任何步骤失败，需要按相反顺序回滚
	wf, err := builder.New("order-saga").
		WithVersion("1.0.0").
		WithDescription("电商订单 Saga 事务").
		NodeWithRollback("create-order",
			func(ctx core.Context) (interface{}, error) {
				orderID := "ORD-001"
				fmt.Printf("✓ 步骤1: 创建订单 %s\n", orderID)
				ctx.Put("order-id", orderID)
				return orderID, nil
			},
			func(ctx core.Context) error {
				orderID, _ := context.GetT[string](ctx, "order-id")
				fmt.Printf("✗ 回滚: 取消订单 %s\n", orderID)
				return nil
			},
		).
		NodeWithRollback("reduce-inventory",
			func(ctx core.Context) (interface{}, error) {
				orderID, _ := context.GetT[string](ctx, "order-id")
				fmt.Printf("✓ 步骤2: 扣减库存 (订单: %s)\n", orderID)
				ctx.Put("inventory-reduced", true)
				return "inventory-reduced", nil
			},
			func(ctx core.Context) error {
				orderID, _ := context.GetT[string](ctx, "order-id")
				fmt.Printf("✗ 回滚: 恢复库存 (订单: %s)\n", orderID)
				return nil
			},
		).
		NodeWithRollback("process-payment",
			func(ctx core.Context) (interface{}, error) {
				orderID, _ := context.GetT[string](ctx, "order-id")
				fmt.Printf("✓ 步骤3: 处理支付 (订单: %s)\n", orderID)
				ctx.Put("payment-id", "PAY-001")
				return "PAY-001", nil
			},
			func(ctx core.Context) error {
				paymentID, _ := context.GetT[string](ctx, "payment-id")
				fmt.Printf("✗ 回滚: 退款 %s\n", paymentID)
				return nil
			},
		).
		NodeWithRollback("ship-item",
			func(ctx core.Context) (interface{}, error) {
				orderID, _ := context.GetT[string](ctx, "order-id")
				fmt.Printf("✓ 步骤4: 发货 (订单: %s)\n", orderID)
				ctx.Put("shipment-id", "SHIP-001")
				return "SHIP-001", nil
			},
			func(ctx core.Context) error {
				shipmentID, _ := context.GetT[string](ctx, "shipment-id")
				fmt.Printf("✗ 回滚: 召回货物 %s\n", shipmentID)
				return nil
			},
		).
		Node("send-notification", func(ctx core.Context) (interface{}, error) {
			// 模拟通知服务失败
			return nil, errors.New("通知服务不可用")
		}).
		DependsOn("reduce-inventory", "create-order").
		DependsOn("process-payment", "reduce-inventory").
		DependsOn("ship-item", "process-payment").
		DependsOn("send-notification", "ship-item").
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
	fmt.Println("\n开始执行 Saga 事务...")
	fmt.Println("────────────────────────────────────")

	result, err := e.Run(ctx)
	if err != nil {
		fmt.Println("────────────────────────────────────")
		fmt.Printf("\n❌ 工作流执行失败: %v\n", err)
		fmt.Println("\n开始执行 Saga 补偿（回滚）...")
		fmt.Println("────────────────────────────────────")

		// 执行回滚
		if rollbackErr := rollbackMgr.Rollback(ctx); rollbackErr != nil {
			fmt.Printf("回滚失败: %v\n", rollbackErr)
		} else {
			fmt.Println("────────────────────────────────────")
			fmt.Println("✅ Saga 补偿完成，所有操作已回滚")
		}
	}

	// 输出结果
	fmt.Println("\n════════════════════════════════════")
	fmt.Println("执行结果:")
	fmt.Printf("  状态: %v\n", result.Status)
	fmt.Printf("  耗时: %v\n", result.Duration)
	fmt.Printf("  执行节点数: %d\n", len(result.NodeResults))

	fmt.Println("\n节点执行详情:")
	for name, nodeResult := range result.NodeResults {
		status := "✓ 成功"
		if nodeResult.Status == core.NodeFailed {
			status = "✗ 失败"
		}
		fmt.Printf("  %s: %s\n", name, status)
	}
}
