package controller

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/zHElEARN/go-csust-planet/utils"
)

var (
	buildingsCache sync.Map
)

// Hello godoc
// @Summary      Hello World测试
// @Description  返回一个简单的hello world消息
// @Tags         util
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /util/hello [get]
func Hello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "hello world",
	})
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
		utils.ResponseError(c, http.StatusBadRequest, "缺少参数: campus, building, room 均为必填")
		return
	}

	var targetCampus utils.Campus
	switch campusName {
	case "云塘":
		targetCampus = utils.CampusYuntang
	case "金盆岭":
		targetCampus = utils.CampusJinpenling
	default:
		utils.ResponseError(c, http.StatusBadRequest, "无效的校区名称")
		return
	}

	var buildings []utils.Building
	if val, ok := buildingsCache.Load(targetCampus.ID); ok {
		buildings = val.([]utils.Building)
	} else {
		var err error
		buildings, err = utils.GetBuildings(targetCampus)
		if err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, "获取楼栋列表失败: "+err.Error())
			return
		}
		buildingsCache.Store(targetCampus.ID, buildings)
	}

	var targetBuilding *utils.Building
	for _, b := range buildings {
		if b.Name == buildingName {
			targetBuilding = &b
			break
		}
	}

	if targetBuilding == nil {
		utils.ResponseError(c, http.StatusNotFound, fmt.Sprintf("在%s未找到楼栋: %s", campusName, buildingName))
		return
	}

	balance, err := utils.GetElectricity(*targetBuilding, roomNum)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "查询电费失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
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
		utils.ResponseError(c, http.StatusBadRequest, "缺少参数: token 不能为空")
		return
	}

	profile, err := utils.GetUserProfile(token)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取用户信息失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profile": profile,
	})
}
