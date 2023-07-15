package handler

import "github.com/gin-gonic/gin"

var router *gin.Engine

func InitRouter(blogHandler *BlogHandler, commentHandler *CommnetHandler) {
	router = gin.Default()

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, "OK")
	})

	blogPath := router.Group("/v1/blog")
	{
		blogPath.POST("/", blogHandler.CreateBlog)
		blogPath.GET("/", blogHandler.GetBlogs)
		blogPath.GET("/:id", blogHandler.GetBlog)
		blogPath.PUT("/:id", blogHandler.UpdateBlog)
	}

	commentPath := router.Group("/v1/comment")
	{
		commentPath.POST("/", commentHandler.CreateComment)
		commentPath.GET("/", commentHandler.GetComments)
		commentPath.GET("/:id", commentHandler.GetComment)
		commentPath.PUT("/:id", commentHandler.UpdateComment)
		commentPath.DELETE("/:id", commentHandler.DeleteComment)
	}
}

func GetRouter(blogHandler *BlogHandler, commentHandler *CommnetHandler) *gin.Engine {
	if router == nil {
		InitRouter(blogHandler, commentHandler)
	}
	return router
}
