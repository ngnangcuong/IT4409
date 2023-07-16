package handler

import (
	"IT4409/internal/app/middleware"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func InitRouter(blogHandler *BlogHandler, commentHandler *CommnetHandler, authHandler *AuthHandler) {
	router = gin.Default()

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, "OK")
	})
	router.Use(middleware.Headers())
	router.Use(middleware.SetupCors())

	authPath := router.Group("/v1/auth")
	{
		authPath.GET("/oauth/google", authHandler.OauthGoogle)
		authPath.POST("/refresh", authHandler.Refresh)
		authPath.GET("/logout", middleware.AuthMiddleware(), authHandler.Logout)
	}

	blogPath := router.Group("/v1/blog")
	{
		blogPath.Use(middleware.AuthMiddleware())
		blogPath.POST("/", blogHandler.CreateBlog)
		blogPath.GET("/", blogHandler.GetBlogs)
		blogPath.GET("/:id", blogHandler.GetBlog)
		blogPath.PUT("/:id", blogHandler.UpdateBlog)
	}

	commentPath := router.Group("/v1/comment")
	{
		commentPath.Use(middleware.AuthMiddleware())
		commentPath.POST("/", commentHandler.CreateComment)
		commentPath.GET("/", commentHandler.GetComments)
		commentPath.GET("/:id", commentHandler.GetComment)
		commentPath.PUT("/:id", commentHandler.UpdateComment)
		commentPath.DELETE("/:id", commentHandler.DeleteComment)
	}
}

func GetRouter(blogHandler *BlogHandler, commentHandler *CommnetHandler, authHandler *AuthHandler) *gin.Engine {
	if router == nil {
		InitRouter(blogHandler, commentHandler, authHandler)
	}
	return router
}
