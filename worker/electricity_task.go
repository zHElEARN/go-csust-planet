package worker

import (
	"log"
	"time"

	"github.com/zHElEARN/go-csust-planet/service"
)

func StartElectricityPushWorker(pushService service.ElectricityPushService, tickInterval time.Duration) {
	if tickInterval <= 0 {
		tickInterval = service.DefaultWorkerTickInterval
	}

	go func() {
		ticker := time.NewTicker(tickInterval)
		defer ticker.Stop()

		log.Println("[INFO] 电费推送 Worker 已启动，等待调度...")

		for range ticker.C {
			if err := pushService.PollAndProcess(time.Now()); err != nil {
				log.Printf("[ERROR] 电费推送 Worker 执行失败: %v", err)
			}
		}
	}()
}
