package campusmap

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/utils/response"
)

type Handler struct{ service *Service }

type featureResponse struct {
	Type       string            `json:"type"`
	Properties FeatureProperties `json:"properties"`
	Geometry   FeatureGeometry   `json:"geometry"`
}

type responseBody struct {
	Type     string            `json:"type"`
	Features []featureResponse `json:"features"`
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

// GetCampusMap godoc
// @Summary 获取校园地图数据
// @Description 获取GeoJSON格式的校园地图数据
// @Tags config
// @Produce json
// @Success 200 {object} responseBody
// @Failure 500 {object} response.ErrorResponse
// @Failure 504 {object} response.ErrorResponse
// @Router /config/campus-map [get]
func (h *Handler) GetCampusMap(c *gin.Context) {
	entities, err := h.service.List(c.Request.Context())
	if err != nil {
		if response.HandleContextError(c, err) {
			return
		}
		log.Printf("[ERROR] 获取校园地图数据失败: %v", err)
		response.ResponseError(c, http.StatusInternalServerError, "获取校园地图数据失败")
		return
	}
	features := make([]featureResponse, 0, len(entities))
	for _, entity := range entities {
		features = append(features, featureResponse{Type: entity.Type, Properties: entity.Properties, Geometry: entity.Geometry})
	}
	c.JSON(http.StatusOK, responseBody{Type: "FeatureCollection", Features: features})
}
