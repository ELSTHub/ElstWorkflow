// Rollback 示例演示了Saga模式回滚的使用
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/elstworkflow/builder"
	"github.com/elstworkflow/context"
	"github.com/elstworkflow/core"
	"github.com/elstworkflow/engine"
	"github.com/elstworkflow/rollback"
)

func main() {
	rollbackMgr := rollback.NewRollbackManager()

	// 构建带回滚的工作流
	wf, err := builder.New("rollback-workflow").
		WithVersion("1.0.0").
		WithDescription("一个带Saga模式回滚的工作流示例").
		NodeWithRollback("create-order",
			func(ctx core.Context) (interface{}, error) {
				fmt.Println("创建订单: ORD-001")
				ctx.Put("order-id", "ORD-001")
				return "ORD-001", nil
			},
			func(ctx core.Context) error {
				orderID, _ := context.GetT[string](ctx, "order-id")
				fmt.Printf("回滚: 取消订单 %s\n", orderID)
				return nil
			},
		).
		NodeWithRollback("process-payment",
			func(ctx core.Context) (interface{}, error) {
				fmt.Println("处理支付: PAY-001")
				ctx.Put("payment-id", "PAY-001")
				return "PAY-001", nil
			},
			func(ctx core.Context) error {
				paymentID, _ := context.GetT[string](ctx, "payment-id")
				fmt.Printf("回滚: 退款 %s\n", paymentID)
				return nil
			},
		).
		NodeWithRollback("ship-item",
			func(ctx core.Context) (interface{}, error) {
				fmt.Println("发货: SHIP-001")
				ctx.Put("shipment-id", "SHIP-001")
				return "SHIP-001", nil
			},
			func(ctx core.Context) error {
				shipmentID, _ := context.GetT[string](ctx, "shipment-id")
				fmt.Printf("回滚: 召回货物 %s\n", shipmentID)
				return nil
			},
		).
		Node("send-notification", func(ctx core.Context) (interface{}, error) {
			// 模拟发送通知失败
			return nil, errors.New("通知服务不可用")
		}).
		DependsOn("process-payment", "create-order").
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
	fmt.Println("开始执行带回滚的工作流...")
	result, err := e.Run(ctx)
	if err != nil {
		fmt.Printf("\n工作流执行失败: %v\n", err)
		fmt.Println("开始执行回滚...")

		// 执行回滚
		if rollbackErr := rollbackMgr.Rollback(ctx); rollbackErr != nil {
			fmt.Printf("回滚失败: %v\n", rollbackErr)
		} else {
			fmt.Println("回滚完成!")
		}
	}

	// 输出结果
	fmt.Printf("\n工作流执行结果:\n")
	fmt.Printf("状态: %v\n", result.Status)
	fmt.Printf("耗时: %v\n", result.Duration)
}
