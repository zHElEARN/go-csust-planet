package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sideshow/apns2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/utils/apns"
	"github.com/zHElEARN/go-csust-planet/utils/campuscard"
)

const (
	WorkerTickInterval  = 1 * time.Minute  // Worker 轮询数据库的间隔时间
	ZombieTaskThreshold = 1 * time.Minute  // 僵尸任务判定阈值（停留在 processing 超过此时间将被重置）
	BatchSizeLimit      = 100              // 每次数据库拉取的最大任务批次数量
	TaskTimeout         = 30 * time.Second // 单个任务的绝对超时时间
)

// 任务状态常量
const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
)

type TaskWithToken struct {
	model.ElectricityTask
	Token string `gorm:"column:device_token"`
}

// 真实的电量查询包装函数
func fetchRealElectricity(campusName, buildingName, roomNum string) (string, error) {
	targetBuilding, err := campuscard.GetBuildingByCampusName(campusName, buildingName)
	if err != nil {
		return "", err
	}

	// 调用底层接口查询电量
	balance, err := campuscard.GetElectricity(targetBuilding, roomNum)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", balance), nil
}

func StartElectricityPushWorker() {
	go func() {
		ticker := time.NewTicker(WorkerTickInterval)
		defer ticker.Stop()

		log.Println("电费推送 Worker 已启动，等待调度...")

		for range ticker.C {
			now := time.Now()

			// 僵尸任务恢复
			// 如果有任务停留在 processing 状态超过设定阈值，说明上次 Worker 崩溃或严重超时，将其重置为 pending
			res := config.DB.Model(&model.ElectricityTask{}).
				Where("status = ? AND updated_at <= ?", TaskStatusProcessing, now.Add(-ZombieTaskThreshold)).
				Update("status", TaskStatusPending)
			if res.Error != nil {
				log.Printf("重置僵尸任务失败: %v\n", res.Error)
			} else if res.RowsAffected > 0 {
				log.Printf("成功将 %d 个僵尸任务重置为 pending\n", res.RowsAffected)
			}

			// 拉取本批次任务
			var tasks []TaskWithToken
			err := config.DB.Transaction(func(tx *gorm.DB) error {
				if err := tx.Table("electricity_tasks").
					Select("electricity_tasks.*, device_tokens.device_token").
					Joins("JOIN device_tokens on electricity_tasks.device_token_id = device_tokens.id").
					Where("electricity_tasks.status = ? AND electricity_tasks.next_run_at <= ?", TaskStatusPending, now).
					Limit(BatchSizeLimit).
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

				// 将这批任务标记为 processing，更新 updated_at 以防被僵尸恢复逻辑误判
				return tx.Model(&model.ElectricityTask{}).
					Where("id IN ?", taskIDs).
					Updates(map[string]any{
						"status":     TaskStatusProcessing,
						"updated_at": now,
					}).Error
			})

			if err != nil {
				log.Printf("获取任务失败: %v\n", err)
				continue
			}

			if len(tasks) == 0 {
				continue
			}

			log.Printf("拉取到 %d 个任务，开始顺序执行...\n", len(tasks))

			// 用来记录本批次中已经判定失效的 DeviceTokenID
			deadTokens := make(map[string]bool)

			// 顺序执行任务
			for _, task := range tasks {
				tokenIDStr := task.DeviceTokenID.String()

				// 如果该设备的 Token 在之前的任务中已经失效并被清理，直接跳过内存中的后续任务
				if deadTokens[tokenIDStr] {
					log.Printf("跳过任务 %v，因为所属设备 Token 已在当前批次中失效并被清理\n", task.ID)
					continue
				}

				// 获取返回值，判断该 Token 是否已经被标记为失效
				tokenInvalidated := processSingleTask(task, now)
				if tokenInvalidated {
					// 加入黑名单，本批次剩下的同 Token 任务会被直接跳过
					deadTokens[tokenIDStr] = true
				}
			}

			log.Printf("本批次 %d 个任务执行完毕\n", len(tasks))
		}
	}()
}

// 处理单个任务
func processSingleTask(task TaskWithToken, batchStartTime time.Time) bool {
	ctx, cancel := context.WithTimeout(context.Background(), TaskTimeout)
	defer cancel()

	// 使用 channel 接收任务执行结果
	errCh := make(chan error, 1)

	go func() {
		// 查询电量
		electricityVal, err := fetchRealElectricity(task.Campus, task.Building, task.Room)
		if err != nil {
			errCh <- fmt.Errorf("获取电量失败: %w", err)
			return
		}

		// 发送 APNs 推送
		notification := apns.PushNotification{
			DeviceToken: task.Token,
			Title:       "宿舍电量通知",
			Body:        fmt.Sprintf("%s%s宿舍当前电量: %s", task.Building, task.Room, electricityVal),
			Sound:       "default",
		}

		errCh <- apns.SendPushNotification(notification)
	}()

	var taskErr error
	select {
	case <-ctx.Done():
		// 任务耗时超过设定阈值，触发超时
		taskErr = fmt.Errorf("任务执行超时(%v)", TaskTimeout)
	case err := <-errCh:
		// 任务在设定时间内执行完毕（成功或报错）
		taskErr = err
	}

	// 根据执行结果处理数据库状态
	if taskErr != nil {
		log.Printf("任务 %v 执行失败: %v\n", task.ID, taskErr)

		reason := taskErr.Error()
		// 识别 APNs 明确告知设备失效的错误
		if reason == apns2.ReasonUnregistered || reason == apns2.ReasonBadDeviceToken {
			log.Printf("检测到设备 Token 失效，正在删除相关的 DeviceToken (ID: %v)\n", task.DeviceTokenID)
			// 直接删除 Token，外键的 OnDelete:CASCADE 会自动清理该设备下的所有 tasks
			config.DB.Where("id = ?", task.DeviceTokenID).Delete(&model.DeviceToken{})

			return true
		} else {
			// 其他错误（网络波动、查询电量超时、学校接口挂了等），将状态回滚为 pending
			config.DB.Model(&model.ElectricityTask{}).
				Where("id = ?", task.ID).
				Updates(map[string]any{
					"status":     TaskStatusPending,
					"updated_at": time.Now(),
				})
		}
	} else {
		// 任务成功，计算下一次通知时间
		log.Printf("任务 %v 执行成功\n", task.ID)

		notifyTimeParsed, _ := time.Parse("15:04", task.NotifyTime)
		nextRunAt := time.Date(
			batchStartTime.Year(), batchStartTime.Month(), batchStartTime.Day(),
			notifyTimeParsed.Hour(), notifyTimeParsed.Minute(), 0, 0, batchStartTime.Location(),
		)

		// 如果计算出的今天通知时间已经过去了，说明是明天的任务
		if !nextRunAt.After(batchStartTime) {
			nextRunAt = nextRunAt.Add(24 * time.Hour)
		}

		// 更新任务为 pending，并设置明天的执行时间
		config.DB.Model(&model.ElectricityTask{}).
			Where("id = ?", task.ID).
			Updates(map[string]any{
				"next_run_at": nextRunAt,
				"status":      TaskStatusPending,
				"updated_at":  time.Now(),
			})
	}
	return false
}
