package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/zHElEARN/go-csust-planet/controller"
	_ "github.com/zHElEARN/go-csust-planet/docs"
	"github.com/zHElEARN/go-csust-planet/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	utilGroup := r.Group("/util")
	utilGroup.Use(middleware.AuthMiddleware())
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

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
