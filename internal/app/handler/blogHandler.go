package handler

import (
	"IT4409/internal/app/models"
	"IT4409/internal/app/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BlogHandler struct {
	blogService *services.BlogService
}

func NewBlogHandler(blogService *services.BlogService) *BlogHandler {
	return &BlogHandler{
		blogService: blogService,
	}
}

func (b *BlogHandler) GetBlog(c *gin.Context) {
	id := c.Param("id")
	successResponse, errorResponse := b.blogService.GetBlog(c, id)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (b *BlogHandler) GetBlogs(c *gin.Context) {
	var getBlogsRequest models.GetBlogsRequest
	if err := c.ShouldBindQuery(&getBlogsRequest); err != nil {
		errorResponse := models.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: models.ErrInvalidParameter.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}
	fmt.Println(getBlogsRequest)
	successResponse, errorResponse := b.blogService.GetBlogs(c, getBlogsRequest)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (b *BlogHandler) CreateBlog(c *gin.Context) {
	var createBlogRequest models.CreateBlogRequest
	if err := c.ShouldBindJSON(&createBlogRequest); err != nil {
		errorResponse := models.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: models.ErrInvalidParameter.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	userID := c.GetString("user_id")
	successResponse, errorResponse := b.blogService.CreateBlog(c, createBlogRequest, userID)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}
	c.JSON(successResponse.Status, successResponse)
}

func (b *BlogHandler) UpdateBlog(c *gin.Context) {
	var updateBlogRequest models.UpdateBlogRequest
	if err := c.ShouldBindJSON(&updateBlogRequest); err != nil {
		errorResponse := models.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: models.ErrInvalidParameter.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}
	id := c.Param("id")
	userID := c.GetString("user_id")
	successResponse, errorResponse := b.blogService.UpdateBlog(c, updateBlogRequest, id, userID)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}
