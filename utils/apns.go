package utils

import (
	"fmt"
	"log"

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

func SendPushNotification(deviceToken string, title string, body string) error {
	if apnsClient == nil {
		return fmt.Errorf("APNS 客户端未初始化")
	}

	p := payload.NewPayload().AlertTitle(title).AlertBody(body).Badge(1).Sound("default")

	notification := &apns2.Notification{}
	notification.DeviceToken = deviceToken
	notification.Topic = config.AppConfig.APNSBundleID
	notification.Payload = p

	res, err := apnsClient.Push(notification)

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
