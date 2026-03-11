package worker

import (
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/utils/apns"
)

type TaskWithToken struct {
	model.ElectricityTask
	Token string `gorm:"column:device_token"`
}

func StartElectricityPushWorker() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		log.Println("电费推送 Worker 已启动，等待每一分钟的调度...")

		for range ticker.C {
			log.Println("触发检查：正在扫描数据库中需要推送的电费任务...")

			now := time.Now()
			var tasks []TaskWithToken

			err := config.DB.Transaction(func(tx *gorm.DB) error {
				if err := tx.Table("electricity_tasks").
					Select("electricity_tasks.*, device_tokens.device_token").
					Joins("JOIN device_tokens on electricity_tasks.device_token_id = device_tokens.id").
					Where("electricity_tasks.status = ? AND electricity_tasks.next_run_at <= ?", "pending", now).
					Clauses(clause.Locking{
						Strength: "UPDATE",
						Table:    clause.Table{Name: "electricity_tasks"},
						Options:  "SKIP LOCKED",
					}).
					Scan(&tasks).Error; err != nil {
					return err
				}

				if len(tasks) == 0 {
					return nil
				}

				var taskIDs []string
				for _, t := range tasks {
					taskIDs = append(taskIDs, t.ID.String())
				}

				return tx.Model(&model.ElectricityTask{}).Where("id IN ?", taskIDs).Update("status", "processing").Error
			})

			if err != nil {
				log.Printf("获取电费任务失败: %v\n", err)
				continue
			}

			if len(tasks) == 0 {
				log.Println("当前没有需要推送的电费任务")
				continue
			}

			for _, t := range tasks {
				// 随便推送一下测试内容
				notification := apns.PushNotification{
					DeviceToken: t.Token,
					Title:       "电费定时推送",
					Body:        "您的电费推送任务已执行，请留意电费情况。",
					Sound:       "default",
				}
				err := apns.SendPushNotification(notification)
				if err != nil {
					log.Printf("任务 %v 推送失败: %v\n", t.ID, err)
				} else {
					log.Printf("任务 %v 推送成功\n", t.ID)
				}

				// 解析 NotifyTime 获取时分
				notifyTimeParsed, err := time.Parse("15:04", t.NotifyTime)
				if err != nil {
					log.Printf("任务 %v 解析 NotifyTime 失败: %v\n", t.ID, err)
					continue
				}

				// 计算下一次执行时间（每天的 NotifyTime）
				nextRunAt := time.Date(now.Year(), now.Month(), now.Day(), notifyTimeParsed.Hour(), notifyTimeParsed.Minute(), 0, 0, now.Location())
				// 如果今天的时间已经过了，那就设置为明天（或者总是加上24小时，只要保证大于当前时间）
				if !nextRunAt.After(now) {
					nextRunAt = nextRunAt.Add(24 * time.Hour)
				}

				// 推送完成后，更新下次运行时间，并将状态恢复为 pending
				config.DB.Model(&model.ElectricityTask{}).Where("id = ?", t.ID).Updates(map[string]interface{}{
					"next_run_at": nextRunAt,
					"status":      "pending",
				})
				log.Printf("任务 %v 下次执行时间已更新为 %v\n", t.ID, nextRunAt.Format(time.RFC3339))
			}
		}
	}()
}
