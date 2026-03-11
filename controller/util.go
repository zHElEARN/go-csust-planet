package controller

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/utils/apns"
	"github.com/zHElEARN/go-csust-planet/utils/campuscard"
	"github.com/zHElEARN/go-csust-planet/utils/response"
	"github.com/zHElEARN/go-csust-planet/utils/sso"
)

var (
	buildingsCache sync.Map
)

type pushRequest struct {
	DeviceToken string `json:"device_token" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Body        string `json:"body" binding:"required"`
}

// Hello godoc
// @Summary      Hello World测试
// @Description  返回一个简单的hello world消息
// @Tags         util
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /util/hello [get]
func Hello(c *gin.Context) {
	response.ResponseSuccess(c, "hello world")
}

// Electricity godoc
// @Summary      查询电费余额
// @Description  根据校区、楼栋和房间号查询对应的电费余额
// @Tags         util
// @Produce      json
// @Param        campus    query     string  true  "校区名称 (例如: 云塘, 金盆岭)"
// @Param        building  query     string  true  "楼栋名称"
// @Param        room      query     string  true  "房间号"
// @Success      200       {object}  map[string]interface{}
// @Failure      400       {object}  map[string]interface{}
// @Failure      404       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]interface{}
// @Router       /util/electricity [get]
func Electricity(c *gin.Context) {
	campusName := c.Query("campus")
	buildingName := c.Query("building")
	roomNum := c.Query("room")

	if campusName == "" || buildingName == "" || roomNum == "" {
		response.ResponseError(c, http.StatusBadRequest, "缺少参数: campus, building, room 均为必填")
		return
	}

	var targetCampus campuscard.Campus
	switch campusName {
	case "云塘":
		targetCampus = campuscard.CampusYuntang
	case "金盆岭":
		targetCampus = campuscard.CampusJinpenling
	default:
		response.ResponseError(c, http.StatusBadRequest, "无效的校区名称")
		return
	}

	var buildings []campuscard.Building
	if val, ok := buildingsCache.Load(targetCampus.ID); ok {
		buildings = val.([]campuscard.Building)
	} else {
		var err error
		buildings, err = campuscard.GetBuildings(targetCampus)
		if err != nil {
			response.ResponseError(c, http.StatusInternalServerError, "获取楼栋列表失败: "+err.Error())
			return
		}
		buildingsCache.Store(targetCampus.ID, buildings)
	}

	var targetBuilding *campuscard.Building
	for _, b := range buildings {
		if b.Name == buildingName {
			targetBuilding = &b
			break
		}
	}

	if targetBuilding == nil {
		response.ResponseError(c, http.StatusNotFound, fmt.Sprintf("在%s未找到楼栋: %s", campusName, buildingName))
		return
	}

	balance, err := campuscard.GetElectricity(*targetBuilding, roomNum)
	if err != nil {
		response.ResponseError(c, http.StatusInternalServerError, "查询电费失败: "+err.Error())
		return
	}

	response.ResponseSuccess(c, "查询电费成功", gin.H{
		"balance":  balance,
		"campus":   campusName,
		"building": buildingName,
		"room":     roomNum,
	})
}

// Profile godoc
// @Summary      获取用户个人信息
// @Description  使用提供的token获取用户档案信息
// @Tags         util
// @Produce      json
// @Param        token  query     string  true  "用户Token"
// @Success      200    {object}  map[string]interface{}
// @Failure      400    {object}  map[string]interface{}
// @Failure      500    {object}  map[string]interface{}
// @Router       /util/profile [get]
func Profile(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.ResponseError(c, http.StatusBadRequest, "缺少参数: token 不能为空")
		return
	}

	profile, err := sso.GetUserProfile(token)
	if err != nil {
		response.ResponseError(c, http.StatusInternalServerError, "获取用户信息失败: "+err.Error())
		return
	}

	response.ResponseSuccess(c, "获取用户信息成功", gin.H{
		"profile": profile,
	})
}

// Push godoc
// @Summary      发送APNS推送消息
// @Description  使用提供的设备Token发送推送通知
// @Tags         util
// @Accept       json
// @Produce      json
// @Param        request  body      pushRequest  true  "推送请求内容"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /util/push [post]
func Push(c *gin.Context) {
	var req pushRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效请求参数: "+err.Error())
		return
	}

	err := apns.SendPushNotification(apns.PushNotification{
		DeviceToken: req.DeviceToken,
		Title:       req.Title,
		Body:        req.Body,
	})
	if err != nil {
		response.ResponseError(c, http.StatusInternalServerError, "推送发送失败: "+err.Error())
		return
	}

	response.ResponseSuccess(c, "推送发送成功")
}
