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
		blogPath.POST("/", middleware.AuthMiddleware(), blogHandler.CreateBlog)
		blogPath.GET("/", blogHandler.GetBlogs)
		blogPath.GET("/:id", blogHandler.GetBlog)
		blogPath.PUT("/:id", middleware.AuthMiddleware(), blogHandler.UpdateBlog)
	}

	commentPath := router.Group("/v1/comment")
	{
		commentPath.POST("/", middleware.AuthMiddleware(), commentHandler.CreateComment)
		commentPath.GET("/", commentHandler.GetComments)
		commentPath.GET("/:id", commentHandler.GetComment)
		commentPath.PUT("/:id", middleware.AuthMiddleware(), commentHandler.UpdateComment)
		commentPath.DELETE("/:id", middleware.AuthMiddleware(), commentHandler.DeleteComment)
	}
}

func GetRouter(blogHandler *BlogHandler, commentHandler *CommnetHandler, authHandler *AuthHandler) *gin.Engine {
	if router == nil {
		InitRouter(blogHandler, commentHandler, authHandler)
	}
	return router
}
