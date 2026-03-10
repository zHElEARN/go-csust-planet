package router

import (
	"github.com/gin-gonic/gin"
	"github.com/zHElEARN/go-csust-planet/controller"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	utilGroup := r.Group("/util")
	{
		utilGroup.GET("/hello", controller.Hello)
		utilGroup.GET("/electricity", controller.Electricity)
		utilGroup.GET("/profile", controller.Profile)
	}

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", controller.Login)
	}

	r.NoRoute(controller.HandleNotFound)

	return r
}
