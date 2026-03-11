package utils

import (
	"fmt"
	"log"
	"time"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
	"github.com/zHElEARN/go-csust-planet/config"
)

var apnsClient *apns2.Client

func InitAPNS() {
	authKey, err := token.AuthKeyFromFile(config.AppConfig.APNSPrivateKeyPath)
	if err != nil {
		log.Fatalf("APNS 令牌加载错误: %v", err)
	}

	jwtToken := &token.Token{
		AuthKey: authKey,
		KeyID:   config.AppConfig.APNSKeyIdentifier,
		TeamID:  config.AppConfig.APNSTeamIdentifier,
	}

	if config.AppConfig.APNSEnvironment == "production" {
		apnsClient = apns2.NewTokenClient(jwtToken).Production()
	} else {
		apnsClient = apns2.NewTokenClient(jwtToken).Development()
	}
}

type PushNotification struct {
	DeviceToken       string         `json:"device_token"`
	Title             string         `json:"title"`
	Subtitle          string         `json:"subtitle"`
	Body              string         `json:"body"`
	Badge             *int           `json:"badge"`     // 使用指针以支持 0 值推送
	Sound             string         `json:"sound"`     // 默认为 "default"
	Category          string         `json:"category"`  // 用于分类通知
	ThreadID          string         `json:"thread_id"` // 用于合并通知
	MutableContent    bool           `json:"mutable_content"`
	ContentAvailable  bool           `json:"content_available"`
	InterruptionLevel string         `json:"interruption_level"` // active, passive, time-sensitive, critical
	Priority          int            `json:"priority"`           // 10 (立即) 或 5 (省电)
	Expiration        int64          `json:"expiration"`         // 过期统一失效时间，0 为永不过期
	CustomData        map[string]any `json:"custom_data"`
}

func SendPushNotification(notification PushNotification) error {
	if apnsClient == nil {
		return fmt.Errorf("APNS 客户端未初始化")
	}

	p := payload.NewPayload().
		AlertTitle(notification.Title).
		AlertBody(notification.Body)

	if notification.Subtitle != "" {
		p.AlertSubtitle(notification.Subtitle)
	}

	if notification.Badge != nil {
		p.Badge(*notification.Badge)
	}

	if notification.Sound != "" {
		p.Sound(notification.Sound)
	} else {
		p.Sound("default")
	}

	if notification.Category != "" {
		p.Category(notification.Category)
	}

	if notification.ThreadID != "" {
		p.ThreadID(notification.ThreadID)
	}

	if notification.MutableContent {
		p.MutableContent()
	}

	if notification.ContentAvailable {
		p.ContentAvailable()
	}

	if notification.InterruptionLevel != "" {
		p.InterruptionLevel(payload.EInterruptionLevel(notification.InterruptionLevel))
	}

	for k, v := range notification.CustomData {
		p.Custom(k, v)
	}

	apnsNotification := &apns2.Notification{}
	apnsNotification.DeviceToken = notification.DeviceToken
	apnsNotification.Topic = config.AppConfig.APNSBundleID
	apnsNotification.Payload = p

	if notification.Priority > 0 {
		apnsNotification.Priority = notification.Priority
	}

	if notification.Expiration > 0 {
		apnsNotification.Expiration = time.Unix(notification.Expiration, 0)
	}

	res, err := apnsClient.Push(apnsNotification)

	if err != nil {
		return err
	}

	if res.Sent() {
		log.Printf("成功推送: %v", res.ApnsID)
		return nil
	} else {
		log.Printf("推送失败: %v %v %v", res.StatusCode, res.ApnsID, res.Reason)
		return fmt.Errorf("推送执行失败: %s", res.Reason)
	}
}
