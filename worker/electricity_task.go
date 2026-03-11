package worker

import (
	"log"
	"time"
)

func StartElectricityPushWorker() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		log.Println("电费推送 Worker 已启动，等待每一分钟的调度...")

		for range ticker.C {
			log.Println("触发检查：正在扫描数据库中需要推送的电费任务...")

			// [TODO]: 接入 GORM 查询和 APNs 推送逻辑
		}
	}()
}
